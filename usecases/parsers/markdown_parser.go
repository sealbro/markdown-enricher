package parsers

import (
	"bufio"
	"bytes"
	"context"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"io"
	"mvdan.cc/xurls/v2"
	"net/http"
	"strings"
)

const GitHubBaseUrl = "https://github.com/"

type MarkdownParser struct {
}

func MakeMarkdownParser() *MarkdownParser {
	return &MarkdownParser{}
}

func (p *MarkdownParser) ExtractLinksFromRemoteFile(ctx context.Context, url string) ([]string, error) {
	buffer, err := getDataFromUrl(ctx, url)
	if err != nil {
		return nil, err
	}

	return p.ExtractLinks(ctx, buffer)
}

func (p *MarkdownParser) ExtractLinks(ctx context.Context, buffer []byte) ([]string, error) {
	reader := text.NewReader(buffer)

	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			extension.NewLinkify(
				extension.WithLinkifyAllowedProtocols([][]byte{
					[]byte("http:"),
					[]byte("https:"),
				}),
				extension.WithLinkifyURLRegexp(
					xurls.Strict(),
				),
			),
		),
	)

	node := markdown.Parser().Parse(reader)

	result := make(map[string]string)
	printer := AstPrinter{}
	printer.visitNodes(result, node, " ")

	var urls []string
	for _, repo := range result {
		urls = append(urls, repo)
	}

	return urls, nil
}

func getDataFromUrl(ctx context.Context, url string) ([]byte, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	io.Copy(writer, res.Body)

	return buffer.Bytes(), nil
}

type AstPrinter struct {
	contents []byte
}

func (v *AstPrinter) visitNodes(result map[string]string, n ast.Node, prefix string) {
	switch node := n.(type) {
	case *ast.Link:
		url := string(node.Destination)

		if strings.HasPrefix(url, GitHubBaseUrl) {
			result[url] = url[len(GitHubBaseUrl):]
		}
	}

	if n.HasChildren() {
		last := n.LastChild()

		for n = n.FirstChild(); ; n = n.NextSibling() {

			v.visitNodes(result, n, prefix+"  ")

			if n == last {
				return
			}
		}
	}
}
