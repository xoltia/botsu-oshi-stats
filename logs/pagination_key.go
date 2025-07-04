package logs

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"time"
)

type PaginationKey struct {
	// Descending order date key.
	Date time.Time
	// Ascending order ID key for tie breaking.
	ID uint64
}

func (k PaginationKey) IsZero() bool {
	return k.ID == 0 && k.Date.IsZero()
}

func (k PaginationKey) EncodeBase64String() string {
	data := make([]byte, 0, 16)
	data = binary.BigEndian.AppendUint64(data, k.ID)
	data = binary.BigEndian.AppendUint64(data, uint64(k.Date.UnixMicro()))
	return base64.RawURLEncoding.EncodeToString(data)
}

func (k *PaginationKey) DecodeBase64String(str string) error {
	data, err := base64.RawURLEncoding.DecodeString(str)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	if len(data) != 16 {
		return fmt.Errorf("decode: invalid length")
	}
	k.ID = binary.BigEndian.Uint64(data[0:8])
	k.Date = time.UnixMicro(int64(binary.BigEndian.Uint64(data[8:16])))
	return nil
}
