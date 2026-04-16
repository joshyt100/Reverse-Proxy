package proxy

//
// import (
// 	"context"
// 	"encoding/hex"
// 	"math/rand"
// 	"net/http"
// 	"strconv"
// 	"time"
// )
//
// type ctxKey string
//
// const requestIDKey ctxKey = "request_id"
//
// func getOrCreateRequestID(r *http.Request) string {
// 	if id := r.Header.Get("X-Request-Id"); id != "" {
// 		return id
// 	}
//
// 	b := make([]byte, 8)
// 	if _, err := rand.Read(b); err != nil {
// 		return strconv.FormatInt(time.Now().UnixNano(), 10)
// 	}
// 	return hex.EncodeToString(b)
// }
//
// func withRequestID(r *http.Request, id string) *http.Request {
// 	ctx := context.WithValue(r.Context(), requestIDKey, id)
// 	return r.WithContext(ctx)
// }
//
// func requestIDFromContext(ctx context.Context) string {
// 	v, _ := ctx.Value(requestIDKey).(string)
// 	return v
// }
