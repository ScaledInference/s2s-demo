package main

import (
	"flag"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	agt "github.com/ScaledInference/amp-go-thin/amp_ai_v2"
	"github.com/mssola/user_agent"
)

const (
	ampAgentUrl        = "http://localhost:8100"
	ampTokenCookieName = "AmpToken"
)

var (
	abTestMode        bool
	irrelevantContext bool

	amp        *agt.Amp
	myTemplate *template.Template
	rng        = rand.New(rand.NewSource(time.Now().UnixNano()))

	candidates = []agt.CandidateField{
		{
			Name: "DonationIncentive",
			Values: []interface{}{
				"stop inhumane hunting of this local treasure",
				"bugs have invaded our community and octopuses feed on them, maintaining a gentle balance",
				"promote tolerance in our society, one animal at a time",
				"lower your taxes while contributing to a cause in our community",
				"tree octopuses are an endangered species, preserve their habitat for the next generation",
			},
		},
	}
)

func main() {
	var (
		useTokens        bool
		projectKey, path string
	)

	flag.StringVar(&projectKey, "key", "", "project key")
	flag.BoolVar(&abTestMode, "abtest", false, "run as an A/B test")
	flag.BoolVar(&irrelevantContext, "irrelevant-context", false, "add irrelevant context variables")
	flag.StringVar(&path, "template-path", filepath.Join(os.Getenv("GOPATH"), "src/github.com/ScaledInference/s2s-demo/web_server/index.html"), "html template path")
	flag.BoolVar(&useTokens, "use_tokens", true, "Use tokens")
	flag.Parse()

	if projectKey == "" {
		panic("Missing project key")
	}

	myTemplate = template.Must(template.ParseFiles(path))

	var err error
	opts := agt.AmpOpts{ProjectKey: projectKey, AmpAgents: []string{ampAgentUrl}, DontUseTokens: !useTokens}
	if amp, err = agt.NewAmp(opts); err != nil {
		panic(err)
	}

	log.Println("Starting AMPED web server. Using project key: ", projectKey)

	if useTokens {
		http.HandleFunc("/", tokenHandler)
	} else {
		http.HandleFunc("/", customHandler)
	}
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func tokenHandler(w http.ResponseWriter, req *http.Request) {
	var err error
	setupHeader(w.Header())

	var ampToken string
	if cookie, err := req.Cookie(ampTokenCookieName); err != nil {
		log.Printf("Can't read cookie: %v", err)
	} else {
		ampToken = cookie.Value
	}

	session, err := amp.CreateNewSession(agt.SessionOpts{AmpToken: ampToken})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Make a contextual decision: pick a donation incentive to show on the web page
	var decision *agt.DecideResponse
	userContext := getContext(req.UserAgent())
	if !abTestMode { // make intelligent decisions using Amp
		if decision, err = session.DecideWithContext("UserContext", userContext, "UserDecision", candidates, 0); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
	} else { // run an A/B test
		if ampToken, err = session.Observe("UserContext", userContext, 0); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		// No decision request is sent out to prevent Amp.ai from learning. So this is an observe only integration
		decision = &agt.DecideResponse{
			Decision: map[string]interface{}{
				"DonationIncentive": candidates[0].Values[rng.Int()%len(candidates[0].Values)],
			},
			AmpToken: ampToken,
		}
	}
	log.Printf("Context is %v, Decision is %v\n", userContext, decision.Decision)

	// Grab the updated token from the decision response and set the cookie
	ampToken = decision.AmpToken
	cookie := &http.Cookie{
		Name:  ampTokenCookieName,
		Value: ampToken,
	}
	http.SetCookie(w, cookie)

	// Debugging
	printDebugInfo(userContext, decision, ampToken)

	// Return the optimized web page
	if err = myTemplate.Execute(w, map[string]interface{}{
		"DonationIncentive": decision.Decision["DonationIncentive"].(string),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func customHandler(w http.ResponseWriter, req *http.Request) {
	var err error
	setupHeader(w.Header())

	var opts agt.SessionOpts
	if userId, ok := req.URL.Query()["user_id"]; ok {
		opts.UserId = userId[0]
	} else {
		http.Error(w, "Didn't get user_id", http.StatusBadRequest)
		log.Println("Error: Didn't get user_id")
		return
	}

	// Construct the amp session object
	session, err := amp.CreateNewSession(opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Make a contextual decision: pick a donation incentive to show on the web page
	var decision *agt.DecideResponse
	userContext := getContext(req.UserAgent())
	if !abTestMode { // make intelligent decisions using Amp
		decision, err = session.DecideWithContext("UserContext", userContext, "UserDecision", candidates, 0)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else { // run an A/B test
		// No decision request is sent out to prevent Amp.ai from learning. So this is an observe only integration
		_, _ = session.Observe("UserContext", userContext, 0)
		decision = &agt.DecideResponse{
			Decision: map[string]interface{}{"DonationIncentive": candidates[0].Values[rng.Int()%len(candidates[0].Values)]},
		}
	}
	log.Printf("Context is %v, Decision is %v\n", userContext, decision.Decision)

	// Return the optimized web page
	if err := myTemplate.Execute(w, map[string]interface{}{
		"DonationIncentive": decision.Decision["DonationIncentive"].(string),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getContext(userAgent string) map[string]interface{} {
	ua := user_agent.New(userAgent)
	log.Println("User-agent = ", userAgent)
	browser, _ := ua.Browser()
	contextProperties := map[string]interface{}{
		"Mobile":   ua.Mobile(),
		"Platform": ua.Platform(),
		"OS":       ua.OS(),
		"Browser":  browser,
	}

	if irrelevantContext {
		contextProperties["JunkX"] = chooseOne("x1", "x2", "x3", "x4")
		contextProperties["JunkY"] = chooseOne("y1", "y2")
		contextProperties["JunkZ"] = chooseOne("z1", "z2", "z3", "z4", "z5", "z6", "z7", "z8")
	}

	return contextProperties
}

func setupHeader(header http.Header) {
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	header.Set("Access-Control-Allow-Headers", "Accept, Accept-Language, Content-Language, Content-Type")
}

func chooseOne(options ...string) string {
	return options[rng.Int()%len(options)]
}

func printDebugInfo(context map[string]interface{}, decision *agt.DecideResponse, ampToken string) {
	log.Printf("Context is %v", context)
	log.Printf("Decision is %v", decision.Decision)
	log.Printf("Returned token is '%v'", ampToken)
	log.Println()
}
