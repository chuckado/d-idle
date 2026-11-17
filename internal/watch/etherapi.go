package watch

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"d2lod-bot/internal/ether"
)

// The ETHER wallet routes (V7, docs/ether.md). GET /api/ether is a read view and mounts with the other GET routes;
// the two POST routes change the wallet and belong on the answer listener, like POST /api/orders.
const (
	EtherRoute     = "/api/ether"
	UpgradeRoute   = "/api/vault/upgrade"
	DowngradeRoute = "/api/vault/downgrade"
	etherSchema    = "watch.ether/1"
)

// EtherAPI serves the wallet. The wallet is account-level, so it is bound once per profile; the vault count the
// refund rule reads is the wallet's own Content hook (the active hero's vault).
type EtherAPI struct {
	W *ether.Wallet
}

// EtherView is GET /api/ether: the wallet summary plus the chip line the Hero sheet and the Vault header draw.
type EtherView struct {
	Schema string `json:"schema"`
	ether.Summary
	Chip string `json:"chip"`
}

type etherReply struct {
	policyReply
	Step     string `json:"step,omitempty"`
	Price    int    `json:"price,omitempty"`
	Balance  int    `json:"balance"`
	Capacity int    `json:"capacity"`
	Need     int    `json:"need,omitempty"`     // 402: how much more ETHER is needed
	Count    int    `json:"count,omitempty"`    // 409 in-use: records in the vault
	Fits     int    `json:"fitsTier,omitempty"` // 409 in-use: records the lower tier holds
	Already  bool   `json:"already,omitempty"`
}

type etherBody struct {
	Step string `json:"step"`
}

// View builds the GET reply.
func (a *EtherAPI) View() EtherView {
	s := a.W.Summary()
	return EtherView{Schema: etherSchema, Summary: s, Chip: chipOf(s)}
}

func chipOf(s ether.Summary) string {
	return s.Currency + " " + strconv.Itoa(s.Balance)
}

// Get is the read route (mount GET only).
func (a *EtherAPI) Get() http.Handler {
	return readOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, a.View())
	}))
}

// Upgrade is POST /api/vault/upgrade, body {"step":"v1"} (empty or missing step: the next one).
func (a *EtherAPI) Upgrade() http.Handler { return a.post(false) }

// Downgrade is POST /api/vault/downgrade: refunds the last step only while the vault's content fits the lower tier.
func (a *EtherAPI) Downgrade() http.Handler { return a.post(true) }

func (a *EtherAPI) post(down bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			writeJSON(w, http.StatusMethodNotAllowed, policyReply{Code: "method", Error: "POST only"})
			return
		}
		var body etherBody
		if b, err := io.ReadAll(io.LimitReader(r.Body, 4<<10)); err != nil {
			writeJSON(w, http.StatusBadRequest, policyReply{Code: "bad-body", Error: err.Error()})
			return
		} else if len(b) > 0 {
			if err := json.Unmarshal(b, &body); err != nil {
				writeJSON(w, http.StatusBadRequest, policyReply{Code: "bad-body", Error: err.Error()})
				return
			}
		}
		var (
			res ether.UpgradeResult
			err error
		)
		if down {
			res, err = a.W.Downgrade(body.Step)
		} else {
			res, err = a.W.Upgrade(body.Step)
		}
		rep := etherReply{}
		code := 200
		var se *ether.ShortError
		var fe *ether.FitError
		switch {
		case err == nil:
			rep.OK, rep.Already = true, res.Already
			rep.Step, rep.Price, rep.Balance, rep.Capacity = res.Step, res.Price, res.Balance, res.Capacity
			rep.Changed = !res.Already
		case errors.As(err, &se):
			code, rep.Code, rep.Need = http.StatusPaymentRequired, "insufficient", se.Need-se.Have
		case errors.As(err, &fe):
			code, rep.Code, rep.Count, rep.Fits = http.StatusConflict, "in-use", fe.Count, fe.Capacity
		case errors.Is(err, ether.ErrOutOfOrder):
			code, rep.Code = http.StatusConflict, "out-of-order"
		case errors.Is(err, ether.ErrNothingToBuy):
			code, rep.Code = http.StatusConflict, "maxed"
		case errors.Is(err, ether.ErrNothingToUndo):
			code, rep.Code = http.StatusConflict, "nothing-to-refund"
		default:
			code, rep.Code = http.StatusInternalServerError, "wallet"
		}
		if err != nil {
			rep.Error = err.Error()
			s := a.W.Summary()
			rep.Balance, rep.Capacity = s.Balance, s.Capacity
		}
		writeJSON(w, code, rep)
	})
}
