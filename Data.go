package zgo

//  Created by Tor Langballe on /31/10/15.

type Data struct {
	buffer []byte
}

func DataFromUrl(surl string, got func(data *Data, err *Error)) {

}

func DataNewFromString(str string) *Data {
	return nil
}

func DataNewFromHex(hex string) *Data {
	return nil
}

func (d *Data) Length() int64 {
	return int64(len(d.buffer))
}

func (d *Data) GetString() string {
	return ""
}

func (d *Data) GetHexString() string {
	return ""
}

func (d *Data) GetBase64() string {
	return ""
}

func (d *Data) SaveToFile(file FilePath) *Error {
	return nil
}

func LoadFromFile(file FilePath) *Error {
	return nil
}
