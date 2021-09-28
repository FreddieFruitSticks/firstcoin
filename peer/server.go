package peer

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
)

// TODO remove all non-server related stuff to a new package - need refactor
type Server struct {
	CoinServerHandler CoinServerHandler
}

func NewServer(cs CoinServerHandler) *Server {
	return &Server{
		CoinServerHandler: cs,
	}
}

func (s *Server) HandleServer(port string) {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong from, %q", html.EscapeString(r.URL.Path))
	})

	http.HandleFunc("/create-block", JSONHandler(s.CoinServerHandler.createBlock)) // control endpoint
	http.HandleFunc("/spend-money", JSONHandler(s.CoinServerHandler.spendMoney))   // control endpoint
	http.HandleFunc("/txpool", JSONHandler(s.CoinServerHandler.getTxPool))         // control endpoint
	http.HandleFunc("/txset", JSONHandler(s.CoinServerHandler.getTxSet))           // control endpoint
	http.HandleFunc("/blockchain", JSONHandler(s.CoinServerHandler.getBlockchain)) // control endpoint

	http.HandleFunc("/block", JSONHandler(s.CoinServerHandler.addBlockToBlockchain))
	http.HandleFunc("/block-chain", JSONHandler(s.CoinServerHandler.blockChain))
	http.HandleFunc("/peers", JSONHandler(s.CoinServerHandler.peers))
	http.HandleFunc("/notify", JSONHandler(s.CoinServerHandler.peers))
	http.HandleFunc("/latest-block", JSONHandler(s.CoinServerHandler.latestBlock))
	http.HandleFunc("/transaction", JSONHandler(s.CoinServerHandler.transaction))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

type ServiceHandler func(*http.Request) (*HTTPResponse, *HTTPError)

func JSONHandler(service ServiceHandler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		httpResponse, err := service(request)

		writer.Header().Set("X-Content-Type-Options", "nosniff")
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")

		if err != nil {
			writer.WriteHeader(err.Code)

			err := json.NewEncoder(writer).Encode(ErrorResponse{
				Type:    "error",
				Message: err.Error(),
			})
			if err != nil {
				fmt.Println(err)
			}

			return
		}

		for key, value := range httpResponse.Headers {
			writer.Header().Set(key, value)
		}

		writer.WriteHeader(httpResponse.StatusCode)
		encodeErr := json.NewEncoder(writer).Encode(httpResponse.Body)
		if encodeErr != nil {
			fmt.Println(encodeErr)
		}
	}
}

type HTTPResponse struct {
	StatusCode int
	Body       interface{}
	Headers    map[string]string
}

type HTTPError struct {
	Code    int
	Message string
}

func (httpError *HTTPError) Error() string {
	return httpError.Message
}

// nolint: unused
func (httpError *HTTPError) ErrorCode() int {
	return httpError.Code
}

func NewHTTPError(code int, message string, args ...interface{}) *HTTPError {
	return &HTTPError{Code: code, Message: fmt.Sprintf(message, args...)}
}

type ErrorResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
