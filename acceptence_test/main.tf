terraform {
  required_version = " >= 0.11.0"
}


provider "ultradns" {
  username = "${var.ULTRADNS_USERNAME}"
  password = "${var.ULTRADNS_PASSWORD}"
  baseurl  = "${var.ULTRADNS_BASEURL}"
}

# Add a record to the domain
resource "ultradns_record" "foobar" {
    name     = "terraform"
    rdata    = [
        "192.168.0.12",
    ]
    ttl      = "3600"
    type     = "A"
    zone     = "kubernetes-ultradns-provider-test.com"

}

resource "ultradns_rdpool" "it" {
  zone        = "${var.ULTRADNS_DOMAINNAME}"
  name        = "test-rdpool-minimal"
  ttl         = 300
  description = "Minimal RD Pool"
  rdata = ["10.6.0.1"]
 
}

resource "ultradns_tcpool" "it" {
  zone        = "${var.ULTRADNS_DOMAINNAME}"
  name        = "${var.ULTRADNS_DOMAINNAME}"
  ttl         = 300
  description = "traffic controller pool with all settings tuned"
  act_on_probes = false
  max_to_lb     = 2
  run_probes    = false

  rdata {
    host = "10.6.1.1"

    failover_delay = 30
    priority       = 1
    run_probes     = true
    state          = "ACTIVE"
    threshold      = 1
    weight         = 2
  }

  backup_record_rdata          = "10.6.1.4"
  backup_record_failover_delay = 30
}

resource "ultradns_record" "it" {
  zone = "${var.ULTRADNS_DOMAINNAME}"
  name  = "test-record-txt"
  rdata = [
    "simple answer",
    "backslash answer \\",
    "quote answer \"",
    "complex answer \\ \"",
  ]
  type  = "TXT"
  ttl   = 3600
}

resource "ultradns_tcpool" "test-probe-ping-pool" {
  zone  = "${var.ULTRADNS_DOMAINNAME}"
  name  = "testprobepingpool"
  ttl   = 30
  description = "traffic controller pool with probes"
  run_probes    = true
  act_on_probes = true
  max_to_lb     = 2

  rdata {
    host = "10.3.0.1"

    state          = "NORMAL"
    run_probes     = true
    priority       = 1
    failover_delay = 0
    threshold      = 1
    weight         = 2
  }

  rdata {
    host = "10.3.0.2"

    state          = "NORMAL"
    run_probes     = true
    priority       = 2
    failover_delay = 0
    threshold      = 1
    weight         = 2
  }

  backup_record_rdata = "10.3.0.3"
}

resource "ultradns_probe_ping" "it" {
  zone  = "${var.ULTRADNS_DOMAINNAME}"
  name  = "testprobepingpool"
  agents = ["DALLAS", "AMSTERDAM"]
  interval  = "ONE_MINUTE"
  threshold = 2

  ping_probe {
    packets    = 15
    packet_size = 56

    limit {
      name     = "lossPercent"
      warning  = 1
      critical = 2
      fail     = 3
    }

    limit {
      name     = "total"
      warning  = 2
      critical = 3
      fail     = 4
    }
  }

  depends_on = ["ultradns_tcpool.test-probe-ping-pool"]
}

resource "ultradns_tcpool" "test-tcp-pool-minimal" {
  zone = "${var.ULTRADNS_DOMAINNAME}"
  name = "testprobehttpminimal.com"
  ttl         = 30
  description = "traffic controller pool with probes"
  run_probes    = true
  act_on_probes = true
  max_to_lb     = 2

  rdata {
    host = "10.2.0.1"

    state          = "NORMAL"
    run_probes     = true
    priority       = 1
    failover_delay = 0
    threshold      = 1
    weight         = 2
  }

  rdata {
    host = "10.2.0.2"

    state          = "NORMAL"
    run_probes     = true
    priority       = 2
    failover_delay = 0
    threshold      = 1
    weight         = 2
  }

  backup_record_rdata = "10.2.0.3"
}

resource "ultradns_probe_http" "test-probe-http-minimal" {
  zone = "${var.ULTRADNS_DOMAINNAME}"
  name = "testprobehttpminimal.com"
  pool_record = "10.2.0.1"
  agents = ["DALLAS", "AMSTERDAM"]
  interval  = "ONE_MINUTE"
  threshold = 2

  http_probe {
    transaction {
      method = "GET"
      url    = "http://www.google.com"

      limit {
        name     = "run"
        warning  = 60
        critical = 60
        fail     = 60
      }

      limit {
        name     = "connect"
        warning  = 20
        critical = 20
        fail     = 20
      }
    }
  }

  depends_on = ["ultradns_tcpool.test-tcp-pool-minimal"]
}

resource "ultradns_tcpool" "test-tcp-pool-maximal" {
  zone  = "${var.ULTRADNS_DOMAINNAME}"
  name  = "testprobehttpmaximal.com"
  ttl   = 30
  description = "traffic controller pool with probes"
  run_probes    = true
  act_on_probes = true
  max_to_lb     = 2

  rdata {
    host = "10.2.1.1"
    state          = "NORMAL"
    run_probes     = true
    priority       = 1
    failover_delay = 0
    threshold      = 1
    weight         = 2
  }

  rdata {
    host = "10.2.1.2"
    state          = "NORMAL"
    run_probes     = true
    priority       = 2
    failover_delay = 0
    threshold      = 1
    weight         = 2
  }

  backup_record_rdata = "10.2.1.3"
}

resource "ultradns_probe_http" "test-probe-http-maximal" {
  zone = "${var.ULTRADNS_DOMAINNAME}"
  name = "testprobehttpmaximal.com"
  pool_record = "10.2.1.1"

  agents = ["DALLAS", "AMSTERDAM"]

  interval  = "ONE_MINUTE"
  threshold = 2

  http_probe {
    transaction {
      method           = "POST"
      url              = "http://www.google.com"
      transmitted_data = "{}"
      follow_redirects = true

      limit {
        name = "run"

        warning  = 1
        critical = 2
        fail     = 3
      }
      limit {
        name = "avgConnect"

        warning  = 4
        critical = 5
        fail     = 6
      }
      limit {
        name = "avgRun"

        warning  = 7
        critical = 8
        fail     = 9
      }
      limit {
        name = "connect"

        warning  = 10
        critical = 11
        fail     = 12
      }
    }

    total_limits {
      warning  = 13
      critical = 14
      fail     = 15
    }
  }

  depends_on = ["ultradns_tcpool.test-tcp-pool-maximal"]
}

resource "ultradns_dirpool" "test-dirpool-minimal" {
  zone        = "${var.ULTRADNS_DOMAINNAME}"
  name        = "testdirpoolminimal.com"
  type        = "A"
  ttl         = 300
  description = "Minimal directional pool"

  rdata {
    host = "10.1.0.1"
    all_non_configured = true
  }
}

resource "ultradns_dirpool" "test-dirpool-maximal" {
  zone        = "${var.ULTRADNS_DOMAINNAME}"
  name        = "testdirpoolmaximal.com"
  type        = "A"
  ttl         = 300
  description = "Description of pool"
  conflict_resolve = "GEO"

  rdata {
    host               = "10.1.1.1"
    all_non_configured = true
  }

  rdata {
    host = "10.1.1.2"

    geo_info {
      name = "North America"

      codes = [
        "US-OK",
        "US-DC",
        "US-MA",
      ]
    }
  }

  rdata {
    host = "10.1.1.3"

    ip_info {
      name = "some Ips"

      ips {
        start = "200.20.0.1"
        end   = "200.20.0.10"
      }

      ips {
        cidr = "20.20.20.0/24"
      }

      ips {
        address = "50.60.70.80"
      }
    }
  }

#   rdata {
#     host = "10.1.1.4"
#
#     geo_info {
#       name             = "accountGeoGroup"
#       is_account_level = true
#     }
#
#     ip_info {
#       name             = "accountIPGroup"
#       is_account_level = true
#     }
#   }

  no_response {
    geo_info {
      name = "nrGeo"

      codes = [
        "Z4",
      ]
    }

    ip_info {
      name = "nrIP"

      ips {
        address = "197.231.41.3"
      }
    }
  }
}

