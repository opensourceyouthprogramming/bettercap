package https_proxy

import (
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"
	"github.com/bettercap/bettercap/tls"
	"github.com/bettercap/bettercap/modules/http_proxy"

	"github.com/evilsocket/islazy/fs"
)

type HttpsProxy struct {
	session.SessionModule
	proxy *http_proxy.HTTPProxy
}

func NewHttpsProxy(s *session.Session) *HttpsProxy {
	p := &HttpsProxy{
		SessionModule: session.NewSessionModule("https.proxy", s),
		proxy:         http_proxy.NewHTTPProxy(s),
	}

	p.AddParam(session.NewIntParameter("https.port",
		"443",
		"HTTPS port to redirect when the proxy is activated."))

	p.AddParam(session.NewStringParameter("https.proxy.address",
		session.ParamIfaceAddress,
		session.IPv4Validator,
		"Address to bind the HTTPS proxy to."))

	p.AddParam(session.NewIntParameter("https.proxy.port",
		"8083",
		"Port to bind the HTTPS proxy to."))

	p.AddParam(session.NewBoolParameter("https.proxy.sslstrip",
		"false",
		"Enable or disable SSL stripping."))

	p.AddParam(session.NewStringParameter("https.proxy.injectjs",
		"",
		"",
		"URL, path or javascript code to inject into every HTML page."))

	p.AddParam(session.NewStringParameter("https.proxy.certificate",
		"~/.bettercap-ca.cert.pem",
		"",
		"HTTPS proxy certification authority TLS certificate file."))

	p.AddParam(session.NewStringParameter("https.proxy.key",
		"~/.bettercap-ca.key.pem",
		"",
		"HTTPS proxy certification authority TLS key file."))

	tls.CertConfigToModule("https.proxy", &p.SessionModule, tls.DefaultSpoofConfig)

	p.AddParam(session.NewStringParameter("https.proxy.script",
		"",
		"",
		"Path of a proxy JS script."))

	p.AddHandler(session.NewModuleHandler("https.proxy on", "",
		"Start HTTPS proxy.",
		func(args []string) error {
			return p.Start()
		}))

	p.AddHandler(session.NewModuleHandler("https.proxy off", "",
		"Stop HTTPS proxy.",
		func(args []string) error {
			return p.Stop()
		}))

	return p
}

func (p *HttpsProxy) Name() string {
	return "https.proxy"
}

func (p *HttpsProxy) Description() string {
	return "A full featured HTTPS proxy that can be used to inject malicious contents into webpages, all HTTPS traffic will be redirected to it."
}

func (p *HttpsProxy) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (p *HttpsProxy) Configure() error {
	var err error
	var address string
	var proxyPort int
	var httpPort int
	var scriptPath string
	var certFile string
	var keyFile string
	var stripSSL bool
	var jsToInject string

	if p.Running() {
		return session.ErrAlreadyStarted
	} else if err, address = p.StringParam("https.proxy.address"); err != nil {
		return err
	} else if err, proxyPort = p.IntParam("https.proxy.port"); err != nil {
		return err
	} else if err, httpPort = p.IntParam("https.port"); err != nil {
		return err
	} else if err, stripSSL = p.BoolParam("https.proxy.sslstrip"); err != nil {
		return err
	} else if err, certFile = p.StringParam("https.proxy.certificate"); err != nil {
		return err
	} else if certFile, err = fs.Expand(certFile); err != nil {
		return err
	} else if err, keyFile = p.StringParam("https.proxy.key"); err != nil {
		return err
	} else if keyFile, err = fs.Expand(keyFile); err != nil {
		return err
	} else if err, scriptPath = p.StringParam("https.proxy.script"); err != nil {
		return err
	} else if err, jsToInject = p.StringParam("https.proxy.injectjs"); err != nil {
		return err
	}

	if !fs.Exists(certFile) || !fs.Exists(keyFile) {
		err, cfg := tls.CertConfigFromModule("https.proxy", p.SessionModule)
		if err != nil {
			return err
		}

		log.Debug("%+v", cfg)
		log.Info("Generating proxy certification authority TLS key to %s", keyFile)
		log.Info("Generating proxy certification authority TLS certificate to %s", certFile)
		if err := tls.Generate(cfg, certFile, keyFile); err != nil {
			return err
		}
	} else {
		log.Info("loading proxy certification authority TLS key from %s", keyFile)
		log.Info("loading proxy certification authority TLS certificate from %s", certFile)
	}

	return p.proxy.ConfigureTLS(address, proxyPort, httpPort, scriptPath, certFile, keyFile, jsToInject, stripSSL)
}

func (p *HttpsProxy) Start() error {
	if err := p.Configure(); err != nil {
		return err
	}

	return p.SetRunning(true, func() {
		p.proxy.Start()
	})
}

func (p *HttpsProxy) Stop() error {
	return p.SetRunning(false, func() {
		p.proxy.Stop()
	})
}
