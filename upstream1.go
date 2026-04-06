//go:build ignore

package main
import "net/http"
func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("upstream1\n"))
    })
    http.ListenAndServe(":9000", nil)
}
