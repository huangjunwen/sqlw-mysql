// Code generated by "esc -o templates.go templates"; DO NOT EDIT.

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDirectory struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	once sync.Once
	data []byte
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDirectory) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		_ = f.Close()
		return b, err
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/templates/default/group.tmpl": {
		local:   "templates/default/group.tmpl",
		size:    830,
		modtime: 1530326882,
		compressed: `
H4sIAAAAAAAC/3RSTWvcMBC961e83UOQycZOr023kEIOhVJCU9JDyEG15XSw9cFYZr1s/N+LLHs/DrkY
S/PmvXlv5FXZqDeNwwH5Y/r/qYzGOApBxjsOkAJY1yasRSZE3dsSb+x6/20vOy7x1FKpeYNOt7oMjhER
kmzQXKtSH8YMv9XfVv9yuz8U/j0yGcX7DQJ3x96JbzlmOAgBUB0huR1+aCszrLa4xfs75IxdbWGpxdXV
3HwBjBQA4JWlUtYm5A/Mjmu5fhi8LgO08WGPLup16ywTwBg1PZPp8HkLoxotjfIvZz5eyYYMRYHZwrNq
e42bryBb6QFumrdI0wigdgyKXLd3IHxBx+Uy4x3o+jq5BChoE2Gp/j1oIymbKoGn+zlXGYGpMH2KAk8N
edi+bcFu1+WJrsYqcP6sWqrkkgNQOhvI9no6jknZR/rA+bkfmRSoGjZwTQRMmbz41yO7a46sRYF7WL27
jKR2va3yGZJ2eO+9tpUMnM3XVONykQslTvucmyy1S9e4dFcDthfP4waf5toyL7YRdu73Q82jYkq/GrJc
zm/xNMYp/oluFGIU4n8AAAD//8prifY+AwAA
`,
	},

	"/templates/default/headnote": {
		local:   "templates/default/headnote",
		size:    103,
		modtime: 1529743910,
		compressed: `
H4sIAAAAAAAC/0TKsRHCMAwF0J4pfgcUWD01I7CAQ4QlzpEJki6X7SnTPyI8RR1v7Qx11IyBxsa/Gjxj
2uFr327L7mvHRSK+fidqGpJTeY2FJKu1T9rGRge9FuAx7BzgWaOc/gEAAP//80UxpWcAAAA=
`,
	},

	"/templates/default/intf.tmpl": {
		local:   "templates/default/intf.tmpl",
		size:    2126,
		modtime: 1531191153,
		compressed: `
H4sIAAAAAAAC/6SVy27rNhCG93qKQTZHSl25L9BFm3YRIA3cJE0XQVHQ0sgiQnEUchTJMPzuBS+65YLi
IDub/Oebf2ZIqhXFszggnE6Q78LvW9EgnM9JIpuWDEOaAFwUpBkHvnC/S8FiLyxu7Yu6SLIk2W7h9wEL
NCAtcI1QUNOQBqkZTSUKBCbAAYuOEQS8dGiO0EuuqWMwyJ3RUh9A6CMY6vOEjy1OxIlxSsAvXgUracED
RFt5XNtEtmUj9WEDwhws5Hk+QU7nDFL7ovI7tJ3iDaAxZLLk7Iv40wV/XxWze0N9arNofiKt3PvVL9u/
9P6pt5P7SL6j/qvwkR0b4to9lvLzOJEfxurmOX1cbtCP7tBE6BWprtH3hdAaDRhsDVrUbEFAQUphwZI0
UOX+dY12sxAMhdCwR7A+rHSZCCpDDYjFmVmj1262W9DDzqDbBdG2qEsLQqk5DUGJlqEiE9JIfcgTmKNS
v3359M+yZcmEJsue3XSWndVCKIUliIpxBvosBJYaBBzYCGgNFWitnbNFUJqF+ca2PYi9wjvq1x1jt/rN
Lnow6dblr1oTTXvpH8giHuNw7Edk4zakrsg0wo3EuZsi0gwupz8R9yiULGeU6RBkBVxL626UgFe/7/HO
L6SYH3LQxHD7181N5vgekWawJ1KRGow/CtXhyqb8kes4O0fuMJ+q9OJUug5ksBjWirgjv/MZ85uFNihm
bgz5kLwe0t+S652RjfBPxDQvVzVVsQXuAYQ2ip7x+GZ+S8R6lKMilhNF7zu0YIcO+dxuHP60hCNQjzFY
RpG07uVrhfEu9zHWuoBlqnTdgFD/vZLFm0s9XjZPs24/Fhq1H9zSG9SrQhTqA9fB/cSAIAw2plt4zdi8
H6l0q6twkGBr6lTpbqqbhNTw9NNmZGYbIK7R9NK6t74VWhbQS6Wc3AhpsQwOXL7PT5oefvG1z++NDl6Y
ll6uq7AqLWipNm5LT/cFm5aPYX92HHijiZAkDYzl23ROkldh/Bf83zePefiM/PZrlmqpsk8FV6T1/0ge
hijIkv8CAAD//wmPM1pOCAAA
`,
	},

	"/templates/default/manifest.json": {
		local:   "templates/default/manifest.json",
		size:    258,
		modtime: 1529741733,
		compressed: `
H4sIAAAAAAAC/2yNSwoCMRBE9zlF0ethDjBXEQlxbH9MPphyITJ3l7TgJ7ireq+SfjhA6hyS572oj6HI
1IHxUnOSoQ1PGvYpU9vmnc0wlkUmtO9aC7vFRhZGk8NLVUbaCUb+COUsEzZWADknHr49IFEZOnS85lvp
GLXyz0v/4Ya3Dljd6p4BAAD//wb8EYkCAQAA
`,
	},

	"/templates/default/meta.tmpl": {
		local:   "templates/default/meta.tmpl",
		size:    11370,
		modtime: 1531191820,
		compressed: `
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
`,
	},

	"/templates/default/meta_test.tmpl": {
		local:   "templates/default/meta_test.tmpl",
		size:    6967,
		modtime: 1531192126,
		compressed: `
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
dTcbAAA=
`,
	},

	"/templates/default/scan_type_map.json": {
		local:   "templates/default/scan_type_map.json",
		size:    622,
		modtime: 1528164482,
		compressed: `
H4sIAAAAAAAC/3zRv2rDMBDH8d1PITQbg1sjiscOBc9NJpEhgSRcUE4QS1PIuwf9udNl8fi1P/yG07NT
Sl+cP4bvLz0rpSxXrzRG54a/2oeerZmkNdOnNVO1J+9dhtnmIviboijA8NNULlJLiqKiYLYWsT00BxhG
w66UmBuN2CNoqeQgS0C6TV0Ut1mwXSYKaKnkIktAumBdFBdcsN0vCmip5CLLAPdzu2GqYZc+Ec5RHwUC
SWX1Gh6AV3b/JYu8rR7b6Jas/+ZN2b26dwAAAP//G8mEhG4CAAA=
`,
	},

	"/templates/default/stmt.tmpl": {
		local:   "templates/default/stmt.tmpl",
		size:    8676,
		modtime: 1530329395,
		compressed: `
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
798BAAD//5V9NJHkIQAA
`,
	},

	"/templates/default/table.tmpl": {
		local:   "templates/default/table.tmpl",
		size:    9881,
		modtime: 1531196957,
		compressed: `
H4sIAAAAAAAC/7xaS3MjtxG+81e0p+w1R0UN40NyUIqVkmU5ZqxQilZyDltbETSDIWENAQoARalY/O+p
RmOGmAcfcjZ7kUi8uvvrJxpcsPSJTTms15Dc0OcJm3PYbHq9npgvlLbQ7wFEqZKWv9oIP+dz+m+sFnJq
3Gcr5tx9yJhlj8zwoXkuoh6OTNXiaZoIOXxRBbOi4Jans6FcFkXy8peoF/d6wyF8XC4WmhsDUllYGp4B
11ppk/RemHYs/AdG4LlIfmTp01Srpcz8RD63yY0W0uZ+wPOW/EMJ6YeQxWSiVv6rY2DCVz8qVSAT6zV8
K1H2sxEkd+yx4PS3xAPnl40F9xVYwyFsF2w2oDmKw6U1wECrFagcLO6Bh4rQZvOQ9Ozbgte3GquXqYV1
D3BcMznl8G2qioDuhSqWc2ngdLOhVThfcYMDH1Mm7/Bot3OzgYffjZJnUbmWTvAbIsged0094Gkip6lf
mPmJ52xZ2N9YscT54RB+YQYyGk3Wa+Ayg5It/Iw8bhy8Ii/5v9FizvRb8hsrBK0YDkG+hjD4JURHGLAz
Dgsagxc3qPIabh7KPaccAWzJ2PsBbgq8RYJkmyyLoqZmYZwJOpt44doIJVsiAe4d2+8NzJmQxRt5Rq40
mJRJKeTU2dZqJtIZzNnbI4fJ/dUV9HkyTUBIUEvLNfyuhIwrfFqcBLgonQmJEJ+EK3rQgBUas0fYKekf
vdtBOSlFJ4Sl56mJMi5rmfJwCD+HEKSOCjlWy4Afaoqp22TDaT8WInW2ZtyHbvtqb/j0uQ4W6ftGc2Qc
2GLBZWaAFYVn1IBVkHFja4pMYDxfFHzuQgZJgPsl1yCk5TpnKU96+VKm0Le6rp94S6/vDj759LnatN7E
TrUnbmbkGeq7r4OjYwzAB6uTpooGdasHiHuV/MpYB0CmUOuzLyGiP7IfU3KANfQANLdLLUGKApB4mS7q
BvtPbhmMYMJXTjj82ndSRUE0jgZu6NNnSh5rWK9P92Oz2eyImoMAFtjQudcL6/f9W9iZj6P9Q/CXnrMj
/h6iX5n8ZhMTGwcCsefUT3gu+sfGy2MY2vKxtZsGW+dLq8YypSPazNWm+0ixc1+dgy6ysa8+KqPw1kTp
hvL1HIeFzJWeM+uDdFcmD43bnXerVsfY9dYgYzjZMrIOTLtlyt7LCJiKZY2ZMgc7EwajGMM8KTLPLOYJ
SgsYgzFLxH+YZUe3H8OjUkXIqNXwzQg90fNHCiBTDYEVp3bmg+H3hrL5H+YloNEXuCGGIPQ57sxK2HQG
opH5xWB/yEuZoXAvYLM5c8ZTydmKhM1A6Asi2rZgUqR9rFEvMW7lzmS3iSUQ4Qy+yzBpo5E5NqMBiDju
AWxqmN4oJ+QeVBe04n/E1dP5ash++L9A66U4BtzFm8NSq9U+aBZv6Kz1iojcAGX+EE6skeUTnBnBidVb
Z9HcHlcX12rYUOFlOfzEg5LYeT+y36V4TDz+uGNsIKTcb2sfY42GkfN4NxAmY4fqjhAWnrsuo/I76vF2
RX4GO6qThu0ElfmpL9XG0nBtMWC64toqEDTi7QDFVhREE5crEOA5tzOVwUoUBbDCKFguMma520OqEBLY
0h2WaqeFskztixz4qzDWxLuRJ6b6qX2t7rwX9H8AHC5fecp1VQVtYSbO725x4wD4AKweQBSVVRmdevla
3qcMkmOmhGAl7Az4q9UM5ioTueAazIKn+ClzslfDWNg9clASbe4MpwBOIRJTqTSPziAaTz5e3t7B+O+T
69tLGE/uriFJkqhaqfmiYKlbent5c3V+0bUoOoMJpt1ie161Zj9yl6/7sRsEErpi71gwy23xcb57ywvF
stC4NI2UxpVrNffGBWNbuTcx4y5MZIMGcrWU2W6LIUrdUj/Dv5Zcv3WajOEFTyspn0nKnBWGx/Vb7LHC
Yoy5UulTp9AUm8jSClz0heUviX8hHLCkejcM9xQKAvGD4EABpZT4Gu/05a3QgZKJPOcaw4ULI4ZCzCP3
Z5AbTq7vLs9qgShTnFpnzj7fSHRpkdSD1Q/DB8lXd/ohgXvDS6O0Ctz/XHMzg4xZRtooO3i7sSYJD3mY
I9nc2sKd5Gp4mdv6buB/4gWvA5/RSIez7ZSNDnl35CVKNTHa/PsUlz8FGe7nX8Muk+a5G8YF3+ZPyW35
vbZkLDP+Giy5l+J5yWm0Woh3xnJtHaeQUJUsf3zD4fxp234JSw1mjEoFGmD3bnh8o+tQ/pT8/GvZeoFc
aS6mEiuU3Ygfx84Bjx5QOMEbSQz9k84zB6S2uLsDiMQ6yozy6t1oWIkcvukoOTzQzWJoUFZEEHQTr0rs
zkZd1QsOUIOp3+qCxfBn2LLIC8MbncoDZ7cLo8BQK8b3KKYyrFA9A3jeiypqoMlgRZcUiD5T4yws1oIL
RukA3onGMhMpN3Xjd4uSsSHvCI2/tD2SRdQECcMH5YR6d740dH86/q3MncaoZ1u5DxbHOxPZIW4OWH0b
btq+D3HXVlW2q7PaUEbNm0ouSydCE7caabZuO1jgajdVJtULJbMqsX6w3lsHEP1wWAI4n/y0q7c7+lvF
clRW+6f7D2zeIK4al4XTsD30VyfIN+ENxws3GoF5LvDSOVG3amX8LNCkd3e6AtUDAde6fjVycFDPpOnN
FZemE5lqhcTpgku/tm3oLo78EWunTvgjXYpzoY0FWRUs7rbJD7kD9IbDcQ4SSYWX7wFulrW+uD+QvLsq
fAzv8padIh1MFFjZ26/jOS2GaykoVTJzmq06zqhSpqd+MLhvuxnfbZFuc60LkrOisDOtltNZ4zot8KhL
uZyjlfwJTn+oTKsyLlxBkJP59M+zDDee/hDv7tcQ79Vrgvs6gGinn0ax2+aEq3bhtzbM8ZY9kcPUItkt
201RuxrJ4X2fWHXBKHgSrhh20SWKey6amWY4czrbEdNMLai1e7Qu0Jky0iG9gRM/SZKuuNKIEVAPEu4k
FySgMz9+tWen9vPhF3p8OtQz3F2K0dNUXQE7X6rC4Oq2dTw8tl61Dj1xzZfGYqhKWVHwDFhuefA8ibAr
MGrOfXNloVXKjTFfDv7Wwxja8wvT7l31QhXe9nvv+UnBntdaX/x2YVcrgEvqI9jxKuRWTZVVcDHj6dPk
/urK2/1hjYKrbDt5OFw579Ewyec6IF6M6sY9ni+UMQKBeNRMpjOXNstfrkQFe+TFVhIM0UKiPcis+llL
RBpKdknec+TO0Wtx3MCCGcMzQqR6nR91mHztKbTXq449I4FOqFXbb1tQ6aMk6LIosFAdgOEWFlajiFIU
xHCdh7LI2dX0bXblS01+ZyBl8nvnMshhNCgtxfkWOtd1SSW8hFak67+RGNumAypZvHkv3DrHITcqSe7q
7PuHn5IJ5JQOpBrppJU30DOvuKTGeXhKwWX/xO2KjztlbPl81xOMP7M88JP4fNyZ5xSJheVzaIVrkYOb
qPX2iUIQxA1VNI3nDveeQoF2/z6kkNQvF9vHmP8GAAD///27flWZJgAA
`,
	},

	"/templates/default/test.tmpl": {
		local:   "templates/default/test.tmpl",
		size:    2130,
		modtime: 1529627540,
		compressed: `
H4sIAAAAAAAC/7RU32/TMBB+jv+KIxKVXZWO3N4m9QEBDzwwKpjGwzQhz3U2a6ljORfKVOV/R3YSmm3p
BGPdw+qzv/vx3d0XJ9WtvNaw3cJ82Z5P5VpD0zBm1q70BJwlab6mlAnGjo7grKIMTAWkK4IzeVXor+UG
NoZugDblG1UW9dqC82Yt/R3c6rs5ozunW7+KfK0ItixR0TD2miUOwVg6Rpa4rD/J/tT0SXE8aVnTeDIc
T6ZwFzhCqaJs2QY4l0WtB267ctzA66f0oSc/dnUM/hbAp4Gp4NYUYoj6buimy/MEaiQWdijBWF5bBZw8
tN5gfy29/qak5avQmOnFpbGkfS6V3jYiUJjGhwVI57Rd8WjOYEJ+rrL212H329kyE4HlSKqyophLgPa+
9CG811R7C9YUg86Ejn7WJGEBp3oTmQWTsyRJw1s6Y0lycdlOZJuqLJ1B6jD+j2eZpU3AfHHUdex93KqK
dwCHqegA3Uvo7gedy7ogvovSg97VVH6yqsXy/mW0pbtyBUz/GAOuPbuxJp3Lwqy4gKuyLIYuHl4t+iY9
9GmLiqvHTdgyAYMhhijVxpC6ARM3WVYa3p6wZBd7rrLuPrt/77C7xwf3Pf74/r3MWLJqexgenLRG8XxN
849h3DlPX6+CBoPiyhy8tNc6nYERgiXNfmbLMtJ5DrfJPnKTfewm++hNDsBv+NXgj5iZPMx9Eece7KFW
kubeOg0DBaTLTto5hfV12Bo4GysDX+YjoFDsCf0PoscnRI+jolcYhf5YiPh3QsRRIeIzhIiHFKLCF1w8
PLSw/rPa3wEAAP//N0x8k1IIAAA=
`,
	},

	"/": {
		isDir: true,
		local: "",
	},

	"/templates": {
		isDir: true,
		local: "templates",
	},

	"/templates/default": {
		isDir: true,
		local: "templates/default",
	},
}
