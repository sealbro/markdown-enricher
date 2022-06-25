job "markdown-enricher" {
  datacenters = ["lan"]

  type = "service"

  constraint {
    attribute = "${attr.cpu.arch}"
    value     = "amd64"
  }

  update {
    max_parallel = 1
    health_check = "checks"
    min_healthy_time = "30s"
    healthy_deadline = "5m"
  }

  migrate {
    max_parallel = 1
    health_check = "checks"
    min_healthy_time = "2m"
    healthy_deadline = "5m"
  }

  group "markdown-enricher-app" {
    count = 1

    network {
      port "app-http" {
        to = 8080
        host_network = "private"
      }
      port "metrics-http" {
        to = 2112
        host_network = "private"
      }
    }

    service {
      name = "markdown-enricher-exporter"
      tags = ["prometheus", "exporter", "infra"]
      port = "metrics-http"

      check {
        type     = "http"
        port     = "metrics-http"
        interval = "30s"
        timeout  = "5s"
        path     = "/healthz"

        check_restart {
          limit = 3
          grace = "90s"
          ignore_warnings = true
        }
      }
    }

    service {
      name = "markdown-enricher"
      tags = ["http", "api", "private", "internal"]
      port = "app-http"

      check {
        type     = "http"
        port     = "metrics-http"
        interval = "30s"
        timeout  = "5s"
        path     = "/healthz"

        check_restart {
          limit = 3
          grace = "90s"
          ignore_warnings = true
        }
      }
    }

    restart {
      attempts = 20
      interval = "30m"
      delay = "1m"
      mode = "fail"
    }

    task "markdown-enricher" {
      driver = "docker"

      config {
        image = "sealbro/markdown-enricher:0.0.2"
        force_pull = true

        ports = ["app-http", "metrics-http"]

        labels {
          from_nomad = "yes"
        }
      }

      template {
        data = <<EOH
GITHUB_TOKEN={{with secret "applications/prod/markdown-enricher"}}{{.Data.data.GITHUB_TOKEN}}{{end}}
EOH

        destination = "secrets/file.env"
        env         = true
      }

      vault {
        policies = ["nomad-server"]
        env = false
      }

      resources {
        cpu    = 300
        memory = 128
      }
    }
  }
}
