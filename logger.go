package dlx

type logger interface {
	Printf(string, ...interface{})
	Tracef(string, ...interface{})
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Noticef(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
}
