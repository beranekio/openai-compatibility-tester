package suites

import "encoding/base64"

// smallPNGBytes returns an 8x8 PNG for multipart image upload tests.
func smallPNGBytes() []byte {
	data, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAgAAAAICAYAAADED76LAAAAEklEQVR4nGP4n2L0Hx9mGBkKACBDpQFoN/xgAAAAAElFTkSuQmCC")
	if err != nil {
		panic(err)
	}
	return data
}