# s2s-demo
Demo for server-side integration with Amp.ai

Pre-requisites
--------------
   1. Linux or Mac system with at least 4GB RAM
   2. Docker 
   3. Golang 1.11
   
Integration Demo 
----------------
   1. Create a demo project in the Amp.ai console at https://amp.ai.
   2. Run Amp-agent locally in a terminal using: `./run-amp-agent.sh <customer_key>`. The customerKey will be provided to you by email. 
   3. After a few minutes, start the web server: `go run web_server/web_server.go --key=<project_key>`. The projectKey can be found under the `INFO` section of the console.
   4. Start the reporting server: `go run reporting_server/reporting_server.go`. 
   5. Open up the Chrome browser and navigate to `http://localhost:8080/`. You'll see a web page asking for a donation for saving the Bay area tree octopus. On the page you'll see a slogan encouraging you to donate. This is the default or control variant in the experiment.
   6. On the web page, enter a donation amount and click "Donate". You'll see the Outcome request in the reporting server logs. The Context and Decision events can be viewed in the web server logs.
   7. If you refresh the page you'll see the same slogan as before, as the web server placed a cookie in your browser. Under the Chrome DeveloperTools "Application" Tab, click the cookies section to inspect the AmpToken. Refreshing the page ensures the same AmpToken is present, guaranteeing decision persistence for the duration of the session. This is useful as a user may return to the page after a while or refresh it, and the outcome event is stitched into the same session using the AmpToken.
   8. In the Amp.ai console, under "INFO", click "View" beside "Total Sessions" to sample sessions. Wait for a couple of minutes if no data is present and retry. Peek into a sample session to view the Context, Decision and Outcome events.
   9. Return to the main page of the console and create a metric "Donation" by clicking on the "+" button on the left panel. Select a suitable metric name and pick the Outcome event "UserOutcome". Choose to maximize existence of the Outcome event by setting 3 up arrows and the "Exists" tab. Thereafter, save the metric.
   10. Next, approve the decision point by clicking on "Decisions" (under the "INFO" section) and checking off the decision. If the decision point is not approved, only the control variant is served by the Amp-agent.
   11. Allocate traffic to the Amped group (or optimized group) by setting "Current Allocation" to 80%. The Amp.ai console tracks statistics for the Amped group of sessions versus the control group continuously.
   12. After a couple of minutes, head back to the browser and purge the AmpToken from the cookie. This is equivalent to loosing a handle to the session and starting a new one. Refresh the page again multiple times and purge the AmpToken each time. You should see a different variant each time, chosen randomly. This confirms that the demo setup is working.
