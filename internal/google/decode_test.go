package google

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func fixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("testdata/shopping.json")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestResponseAndWholeHour(t *testing.T) {
	body, _ := json.Marshal([]interface{}{[]interface{}{"di", 10}, []interface{}{"wrb.fr", nil, fixture(t)}})
	flights, err := ParseResponse(append([]byte(")]}'\n"), body...))
	if err != nil {
		t.Fatal(err)
	}
	if len(flights) != 1 || flights[0].Price != 733 || flights[0].TotalDuration != 2355 {
		t.Fatalf("bad flights: %+v", flights)
	}
	leg := flights[0].Legs[0]
	if leg.DepTime != [2]int{13, 0} || leg.ArrDate != [3]int{2026, 10, 7} {
		t.Fatalf("bad leg: %+v", leg)
	}
	if toIntPair([]interface{}{nil, float64(45)}) != [2]int{0, 45} {
		t.Fatal("midnight minute lost")
	}
}

func TestResponseErrors(t *testing.T) {
	for _, body := range []string{"", "[]", "null", "[null]", `[["wrb.fr",null,null,null,null,[13]]]`, `[["wrb.fr",null,"null"]]`, `<html>consent</html>`} {
		if _, err := ParseResponse([]byte(body)); err == nil {
			t.Errorf("accepted malformed response %q", body)
		}
	}
	_, err := ParseResponse([]byte(`[["wrb.fr",null,null,null,null,[13]]]`))
	if !strings.Contains(err.Error(), "RPC error: [13]") {
		t.Fatal(err)
	}
	flights, err := ParseResponse([]byte(`[["wrb.fr",null,"[null,null,null,null]"]]`))
	if err != nil || len(flights) != 0 {
		t.Fatalf("empty result: %v %v", flights, err)
	}
}

func TestPageMatchesRequestedSearch(t *testing.T) {
	context := `[null,[2,[null,null,2,null,[],1,[1,0,0,0],null,null,null,null,null,null,[[[[["SFO",0]]],[[["SUB",0]]],null,0,null,null,"2026-10-06",null,null,null,null,null,null,null,3]]]]]`
	page := "<script>AF_initDataCallback({key: 'ds:0', data:" + context + ", sideChannel: {}});AF_initDataCallback({key: 'ds:1', data:" + fixture(t) + ", sideChannel: {}});</script>"
	flights, err := parsePage([]byte(page), "SFO", "SUB", "2026-10-06")
	if err != nil || len(flights) != 1 {
		t.Fatalf("page failed: %v %v", flights, err)
	}
	for _, bad := range []string{strings.Replace(page, "2026-10-06", "2026-10-07", 1), strings.Replace(page, `"SFO"`, `"LAX"`, 1), strings.Replace(page, `null,null,2,null`, `null,null,1,null`, 1), "<html>consent</html>"} {
		if _, err := parsePage([]byte(bad), "SFO", "SUB", "2026-10-06"); err == nil {
			t.Fatal("accepted mismatched/missing search")
		}
	}
}
