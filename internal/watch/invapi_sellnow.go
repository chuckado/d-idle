package watch

import "net/http"

// RouteSellNow is the schema 10 route of the inventory page: queue the sell-now errand (docs/inventory-actions.md).
// POST on the answer listener with the gates of every other inventory route; the body is `{}` and unknown fields are
// refused by the policy gate.
const RouteSellNow = "/api/inventory/sellnow"

// InvSellDesk is the optional schema 10 part of the desk. A desk without it answers the route with code "off".
type InvSellDesk interface {
	RequestSell() (any, error)
}

// sellNow serves the route: ok is false when the desk has no sale step (the caller replies code off).
func sellNow(d InvDesk) (res any, err error, ok bool) {
	sd, ok := d.(InvSellDesk)
	if !ok {
		return nil, nil, false
	}
	res, err = sd.RequestSell()
	return res, err, true
}

func sellNowOff(w http.ResponseWriter) {
	writeJSON(w, http.StatusConflict, policyReply{Code: "off", Error: "selling on request is not available on this bot"})
}
