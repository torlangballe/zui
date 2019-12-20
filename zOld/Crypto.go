package zui

//  Created by Tor Langballe on 06/02/2019.

/*
import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"hash/crc32"

	"github.com/torlangballe/zutil/ustr"
)

func CryptoHmacSha1ToBase64(data *Data, key string) string {
	keyForSign := []byte(key)
	h := hmac.New(sha1.New, keyForSign)
	h.Write(data.buffer)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// static func Sha1AsHex(_ data:ZData) -> String {
//     var result = [UInt8](repeating:0, count:Int(CC_SHA1_DIGEST_LENGTH))
//     CC_SHA1((data as NSData).bytes, CC_LONG(data.count), &result)
//     let data = Data(bytes:result)
//     return (data as ZData).GetHexString()
// }

func CryptoMakeUuid() string {
	return ustr.GenerateUUID()
	// return fmt.Sprintf("%x", uuid.NewV4()) this gives weird results
}

var crc32q = crc32.MakeTable(0xD5828281)

// CryptoMD5ToHex returns an md5 hash for all bytes in data as a hex string
func CryptoMD5ToHex(data *Data) string {
	return fmt.Sprintf("%x", md5.Sum(data.buffer))
}

// CryptoCrc32 makes a CRC-32 checksum of bytes.
// A uses a table generated using 0xD5828281
func CryptoCrc32(crc uint32, bytes []byte) uint32 {
	return crc32.Checksum(bytes, crc32q)
}
*/
