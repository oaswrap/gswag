package basicdata

import (
	"encoding/json"
	"net/http"
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
		writeJSON(w, http.StatusOK, u)
	})

	r.HandleFunc("POST /basicdata", func(w http.ResponseWriter, r *http.Request) {
		var u AllBasicDataTypes
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, u)
	})

	r.HandleFunc("GET /basicdata-pointers", func(w http.ResponseWriter, r *http.Request) {
		u := AllBasicDataTypesPointers{
			Int:     ptr(1),
			Int8:    ptr(int8(2)),
			Int16:   ptr(int16(3)),
			Int32:   ptr(int32(4)),
			Int64:   ptr(int64(5)),
			Uint:    ptr(uint(6)),
			Uint8:   ptr(uint8(7)),
			Uint16:  ptr(uint16(8)),
			Uint32:  ptr(uint32(9)),
			Uint64:  ptr(uint64(10)),
			Float32: ptr(float32(1.23)),
			Float64: ptr(float64(4.56)),
			Byte:    ptr(byte('a')),
			Rune:    ptr(rune('b')),
			String:  ptr("test"),
			Bool:    ptr(true),
		}
		writeJSON(w, http.StatusOK, u)
	})

	r.HandleFunc("POST /basicdata-pointers", func(w http.ResponseWriter, r *http.Request) {
		var u AllBasicDataTypesPointers
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, u)
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// Ignoring error since this is just test data.
	_ = json.NewEncoder(w).Encode(v)
}

func ptr[T any](v T) *T {
	return &v
}
