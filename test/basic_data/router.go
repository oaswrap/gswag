package basicdata

import (
	"encoding/json"
	"net/http"

	"github.com/oaswrap/gswag/test/util"
)

type AllBasicDataTypes struct {
	Int     int     `json:"int"`
	Int8    int8    `json:"int8"`
	Int16   int16   `json:"int16"`
	Int32   int32   `json:"int32"`
	Int64   int64   `json:"int64"`
	Uint    uint    `json:"uint"`
	Uint8   uint8   `json:"uint8"`
	Uint16  uint16  `json:"uint16"`
	Uint32  uint32  `json:"uint32"`
	Uint64  uint64  `json:"uint64"`
	Float32 float32 `json:"float32"`
	Float64 float64 `json:"float64"`
	Byte    byte    `json:"byte"`
	Rune    rune    `json:"rune"`
	String  string  `json:"string"`
	Bool    bool    `json:"bool"`
}

type AllBasicDataTypesPointers struct {
	Int     *int     `json:"int"`
	Int8    *int8    `json:"int8"`
	Int16   *int16   `json:"int16"`
	Int32   *int32   `json:"int32"`
	Int64   *int64   `json:"int64"`
	Uint    *uint    `json:"uint"`
	Uint8   *uint8   `json:"uint8"`
	Uint16  *uint16  `json:"uint16"`
	Uint32  *uint32  `json:"uint32"`
	Uint64  *uint64  `json:"uint64"`
	Float32 *float32 `json:"float32"`
	Float64 *float64 `json:"float64"`
	Byte    *byte    `json:"byte"`
	Rune    *rune    `json:"rune"`
	String  *string  `json:"string"`
	Bool    *bool    `json:"bool"`
}

func NewRouter() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("GET /basicdata", func(w http.ResponseWriter, r *http.Request) {
		u := AllBasicDataTypes{
			Int:     1,
			Int8:    2,
			Int16:   3,
			Int32:   4,
			Int64:   5,
			Uint:    6,
			Uint8:   7,
			Uint16:  8,
			Uint32:  9,
			Uint64:  10,
			Float32: 1.23,
			Float64: 4.56,
			Byte:    'a',
			Rune:    'b',
			String:  "test",
			Bool:    true,
		}
		util.WriteJSON(w, http.StatusOK, u)
	})

	r.HandleFunc("POST /basicdata", func(w http.ResponseWriter, r *http.Request) {
		var u AllBasicDataTypes
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			util.WriteErrorJSON(w, http.StatusBadRequest, "invalid input")
			return
		}
		util.WriteJSON(w, http.StatusOK, u)
	})

	r.HandleFunc("GET /basicdata-pointers", func(w http.ResponseWriter, r *http.Request) {
		u := AllBasicDataTypesPointers{
			Int:     util.Ptr(1),
			Int8:    util.Ptr(int8(2)),
			Int16:   util.Ptr(int16(3)),
			Int32:   util.Ptr(int32(4)),
			Int64:   util.Ptr(int64(5)),
			Uint:    util.Ptr(uint(6)),
			Uint8:   util.Ptr(uint8(7)),
			Uint16:  util.Ptr(uint16(8)),
			Uint32:  util.Ptr(uint32(9)),
			Uint64:  util.Ptr(uint64(10)),
			Float32: util.Ptr(float32(1.23)),
			Float64: util.Ptr(float64(4.56)),
			Byte:    util.Ptr(byte('a')),
			Rune:    util.Ptr(rune('b')),
			String:  util.Ptr("test"),
			Bool:    util.Ptr(true),
		}
		util.WriteJSON(w, http.StatusOK, u)
	})

	r.HandleFunc("POST /basicdata-pointers", func(w http.ResponseWriter, r *http.Request) {
		var u AllBasicDataTypesPointers
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			util.WriteErrorJSON(w, http.StatusBadRequest, "invalid input")
			return
		}
		util.WriteJSON(w, http.StatusOK, u)
	})

	return r
}
