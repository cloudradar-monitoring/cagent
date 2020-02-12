package mysql

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/monitoring"
)

const (
	defaultPort = "3306"
)

var log = logrus.WithField("package", "mysql")

type Config struct {
	Enabled        bool    `toml:"enabled" comment:"Set 'false' to disable checking available updates"`
	Connect        string  `toml:"connect" comment:"Use 127.0.0.1 or the path to the mysql.socket, connecting to remote databases is not supported"`
	User           string  `toml:"user" comment:"Create a user with minimal rights\nmysql > GRANT USAGE ON *.* TO cagent@localhost IDENTIFIED BY '<password>';"`
	Password       string  `toml:"password" comment:"Maximum time the package manager is allowed to spend fetching available updates"`
	ConnectTimeout float64 `toml:"connect_timeout" comment:"Maximum time to wait for mysql server to connect using provided credentials"`
}

func (cfg *Config) Validate() error {
	if cfg.ConnectTimeout < 0 {
		return fmt.Errorf("connect_timeout should be equal or greater than 0.0")
	}
	return nil
}

func CreateModule(config *Config) monitoring.Module {
	return &Mysql{
		config: config,
	}
}

type Mysql struct {
	config *Config
	client *sql.DB

	lastStatus     *Status
	lastStatusTime time.Time
}

func (r *Mysql) getClient() (*sql.DB, error) {
	if r.client != nil {
		return r.client, nil
	}

	if len(r.config.Connect) == 0 {
		return nil, fmt.Errorf("connect address is empty")
	}

	if len(r.config.User) == 0 {
		return nil, fmt.Errorf("user is empty")
	}

	var host, port string
	portDelimiter := strings.Index(r.config.Connect, ":")
	if portDelimiter > 0 {
		if strings.Count(r.config.Connect, ":") > 1 {
			// probably ipv6
			return nil, fmt.Errorf("connect: IPv6 address not supported")
		}
		host = r.config.Connect[0:portDelimiter]
		if portDelimiter+1 < len(r.config.Connect) {
			port = r.config.Connect[portDelimiter+1:]
		} else {
			port = defaultPort
		}
	} else {
		host = r.config.Connect
		port = defaultPort
	}

	if host == "127.0.0.1" || host == "localhost" {
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/information_schema?timeout=%.2fs", r.config.User, r.config.Password, host, port, r.config.ConnectTimeout))
		if err != nil {
			return nil, fmt.Errorf("failed to connect %s: %s", r.config.Connect, err.Error())
		}

		r.client = db
		return db, nil
	}

	ip := net.ParseIP(host)
	if ip != nil {
		return nil, fmt.Errorf("connect: only local mysql server is supported. Please use 127.0.0.1, localhost or unix socket path")

	}
	// this is probably unix socket filepath
	absPath, err := filepath.Abs(r.config.Connect)
	if err != nil {
		return nil, fmt.Errorf("connect: not a unix socket path: %s", err.Error())
	}

	if _, err := os.Stat(absPath); err == os.ErrNotExist {
		return nil, fmt.Errorf("connect: provided unix socket path not found")
	} else if err != nil {
		return nil, fmt.Errorf("connect: provided unix socket not valid, %s", err.Error())
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@unix(%s)/information_schema?timeout=%.2fs", r.config.User, r.config.Password, absPath, r.config.ConnectTimeout))
	if err != nil {
		return nil, fmt.Errorf("failed to connect %s: %s", r.config.Connect, err.Error())
	}

	r.client = db
	return db, nil
}

func (r *Mysql) GetDescription() string {
	return fmt.Sprintf("MySQL/MariaDB performance for %s", r.config.Connect)
}

func (r *Mysql) IsEnabled() bool {
	return r.config.Enabled
}

func (r *Mysql) Run() ([]*monitoring.ModuleReport, error) {
	report := monitoring.NewReport(
		fmt.Sprintf("MySQL/MariaDB performance metrics for %s", r.config.Connect),
		time.Now(),
		"",
	)

	client, err := r.getClient()
	if err != nil {
		report.AddAlert(err.Error())
		return []*monitoring.ModuleReport{&report}, nil
	}

	statusTime := time.Now()
	status, err := getStatus(client)
	if err != nil {
		report.AddAlert(fmt.Sprintf("failed to get status: %s", err.Error()))
		return []*monitoring.ModuleReport{&report}, nil
	}

	if r.lastStatus == nil {
		r.lastStatus = status
		r.lastStatusTime = statusTime

		// need one more iteration to calculate
		// but we need to provide nil measurements for consistency
		report.Measurements = emptyResults()
		return []*monitoring.ModuleReport{&report}, nil
	}

	report.Measurements = make(map[string]interface{})
	fillResultsPerSecond(status, r.lastStatus, statusTime.Sub(r.lastStatusTime), report.Measurements)
	return []*monitoring.ModuleReport{&report}, nil
}
