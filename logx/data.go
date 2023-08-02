package logx

type Logger interface {
	Error(msg string, data ...Data)
	Info(msg string, data ...Data)
	Debug(msg string, data ...Data)
	Sync() error
}

type Data struct {
	key   string
	value any
}

func newData(key string, value any) Data {
	return Data{
		key:   key,
		value: value,
	}
}

func S(key string, value string) Data {
	return newData(key, value)
}

func SS(key string, value []string) Data {
	return newData(key, value)
}

func B(key string, value bool) Data {
	return newData(key, value)
}

func Err(err error) Data {
	return newData("err", err)
}
