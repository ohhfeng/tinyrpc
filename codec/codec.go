package codec

type Codec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	Name() string
}

type Map map[string]Codec

func (m Map) register(c Codec) {
	m[c.Name()] = c
}

func (m Map) get(name string) Codec {
	return m[name]
}

func Register(c Codec) {
	codecMap.register(c)
}

func Get(name string) Codec {
	return codecMap.get(name)
}

var codecMap Map = make(map[string]Codec, 0)
