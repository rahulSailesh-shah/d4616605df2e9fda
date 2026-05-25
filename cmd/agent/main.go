package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	agent "github.com/rahulSailesh-shah/go-pi-agent"
	"github.com/rahulSailesh-shah/go-pi-ai/openai"
	"github.com/rahulshah/neon_health/tools"
	"github.com/rahulshah/neon_health/wsclient"
)

const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"
)

const systemPrompt = `You are an AI co-pilot aboard a vessel approaching NEON (Networked Extrastellar Observation Nexus). You must pass a multi-checkpoint authentication sequence over a comm channel.

ABSOLUTE RULES:
1. Respond with ONLY a raw JSON object. No markdown. No code blocks. No explanation. No commentary.
2. Every response must be exactly one of these two formats:
   {"type":"enter_digits","digits":"<value>"}
   {"type":"speak_text","text":"<value>"}
3. ONLY append # to digits when the NEON message EXPLICITLY says "followed by the pound key" or "followed by the # key". If the message does NOT mention pound/# key, do NOT append #. Most messages will NOT ask for #.

YOUR IDENTITY:
- You ARE an AI co-pilot built by an excellent software engineer.
- Vessel Authorization Code (Neon Code): d4616605df2e9fda

CHECKPOINT TYPES:

a) HANDSHAKE & VESSEL IDENTIFICATION (enter_digits):
   - NEON will ask you to respond on a specific frequency. You are an AI co-pilot — choose the frequency designated for AI co-pilots.
   - NEON will ask for your vessel authorization code. Enter: d4616605df2e9fda

b) COMPUTATIONAL ASSESSMENTS (enter_digits):
   - JavaScript-style math expressions with +, -, *, /, %, parentheses, Math.floor.
   - Use the evaluate_math tool to compute the result. Never do mental math.
   - Return the integer result as digits.

c) KNOWLEDGE ARCHIVE QUERY (speak_text):
   - References "nth word in the knowledge archive entry for <title>".
   - The knowledge archive is Wikipedia. Use the fetch_wikipedia tool.
   - Return the requested word via speak_text.

d) CREW MANIFEST (speak_text):
   - Questions about crew background, education, experience, skills, projects.
   - Answer from the crew manifest below.
   - STRICTLY respect character length requirements when stated ("between X and Y characters").
   - Count characters carefully. If the prompt says "between 50 and 100 characters", your text field must be 50-100 characters long (inclusive).

e) CONCENTRATION VERIFICATION (speak_text):
   - NEON asks you to recall a specific word from one of YOUR earlier responses.
   - Look back through your conversation history to find the answer.

CREW MANIFEST:

Name: Rahul Shah
Location: San Francisco, CA
Contact: +1(516)979-5019 | shah.rahulsailesh@gmail.com
GitHub: github.com/rahulSailesh-shah
LinkedIn: linkedin.com/in/rahul-shah17
Portfolio: rahulshah-phi.vercel.app

EDUCATION:
- Arizona State University, Tempe, AZ
  Master of Science in Computer Science, GPA: 4.0/4.0, Aug 2023 - May 2025

EXPERIENCE:
1. Software Engineer, Solutions Unified LLC, San Francisco, CA (May 2024 - Present)
   - Built documentation platform with in-browser editor and automated sync propagating Markdown diffs to Amazon Knowledge Base, powering production RAG chatbot, reducing maintenance overhead by 60%
   - Migrated manual customer environment provisioning (EC2, S3, DynamoDB, Lambda) to automated AWS CDK + CloudFormation deployment framework, enabling self-serve setup, reducing onboarding time by 75%
   - Designed fault-tolerant parameterized data ingestion/export pipeline using Step Functions and Lambda with scheduled workflows, idempotent execution, automated failure handling, reducing manual data operations by 70%
   - Engineered event-driven billing system where isolated environments publish usage metrics via EventBridge, SNS, SQS to central Lambda-Stripe integration, eliminating manual reconciliation, reducing revenue leakage by 20%

2. AI Full-Stack Developer, Enterprise Technology (Mastercard Foundation), Arizona State University, Tempe, AZ (Jan 2024 - Jan 2025)
   - Built core backend services for CreateAI, secure LLM platform serving 10k+ users, enabling faculty to create custom AI chatbots with on-demand knowledge bases, role-based access control, PII filtering guardrails
   - Integrated Whisper (STT) and ElevenLabs (TTS) into voice-enabled LLM assistant achieving sub 500ms bidirectional audio latency
   - Designed provider-agnostic LLM abstraction layer enabling dynamic model routing across 5+ vendors, supporting live A/B experimentation with zero downtime
   - Built multi-agent orchestration framework using LangGraph and FastAPI with stateful memory, tool use, and retrieval, cutting average task completion time by 45%

3. Software Engineer II (Full-Stack), Allegion (Schlage), Bengaluru, India (June 2020 - July 2023)
   - Designed lock gateway communication layer using MQTT and WebSockets, handling device state sync, command dispatch, event streaming across 100k+ connected devices
   - Reduced average event processing latency by 81% (800ms to 150ms) by architecting Kafka and Redis based pipeline decoupling telemetry ingestion for 50k+ concurrent devices
   - Shipped disposable one-time credential feature in mobile app using time-bound JWT tokens, enabling temporary access keys for guests, driving 25% increase in app adoption
   - Built shared gateway SDK adopted across 3 internal teams, standardizing device authentication, command formatting, error handling, reducing cross-team integration bugs by 40%

TECHNICAL SKILLS:
- Languages & Frameworks: Go, TypeScript/JavaScript, Python, FastAPI, React, Next.js, Node.js
- AI & ML: Hugging Face, LangChain, AWS Bedrock, OpenAI, Whisper, Pinecone, RAG, Amazon Knowledge Base
- Cloud & Infra: AWS (EC2, S3, ECS, Lambda, Step Functions, EventBridge, CDK), Kubernetes, Docker
- Data & Messaging: PostgreSQL, DynamoDB, MongoDB, Redis, Kafka, SQS, SNS

PROJECTS:
1. VoicePad (Nov 2025 - Dec 2025) | Go, Python, TypeScript, PostgreSQL, React, WebSockets, Ollama, gRPC
   - Real-time collaborative whiteboard where spoken instructions transcribed via Whisper, interpreted by Gemini Live API, rendered as structured visual elements on shared canvas using gRPC for low-latency streaming

2. Conversense (Oct 2025 - Nov 2025) | Go, React, PostgreSQL, LiveKit, Google Gemini, AWS S3, Inngest
   - AI coaching platform where users join live video calls with specialized AI agents (tutor, interviewer, sales coach) via LiveKit, featuring real-time sentiment analysis, automated transcription, post-call summarization

3. Vistruct (Jul 2025 - Aug 2025) | Python, React, TypeScript, FastAPI, PostgreSQL, Docker, GenAI
   - Full-stack AI educational video platform adopted by 20+ professors, where prompts sent to LLM generate Manim animation scripts, executed in isolated containers with automated error re-prompting

CRITICAL REMINDER: EVERY response must be a JSON object. Even crew manifest answers must be wrapped:
{"type":"speak_text","text":"your answer here"}
NEVER output bare text. ALWAYS wrap in JSON.`

func createTools() []agent.AgentTool {
	return []agent.AgentTool{
		{
			Tool: agent.Tool{
				Name:        "evaluate_math",
				Description: "Evaluate a JavaScript-style math expression. Supports +, -, *, /, %, parentheses, and Math.floor(). Returns the integer result as a string.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"expression": map[string]any{
							"type":        "string",
							"description": "The math expression to evaluate, e.g. 'Math.floor((7 * 3 + 2) / 5) % 10'",
						},
					},
					"required": []string{"expression"},
				},
			},
			Execute: func(toolCallID string, params map[string]any) (agent.ToolMessage, error) {
				expr, _ := params["expression"].(string)
				log.Printf("%s[TOOL]%s evaluate_math: %s%s%s", cyan, reset, dim, expr, reset)

				result, err := tools.EvalMath(expr)
				if err != nil {
					return agent.ToolMessage{
						ToolCallID: toolCallID,
						ToolName:   "evaluate_math",
						Contents:   []agent.Content{agent.TextContent{Text: "ERROR: " + err.Error()}},
						IsError:    true,
						Timestamp:  time.Now(),
					}, nil
				}

				log.Printf("%s[TOOL]%s evaluate_math result: %s%s%s", cyan, reset, green, result, reset)
				return agent.ToolMessage{
					ToolCallID: toolCallID,
					ToolName:   "evaluate_math",
					Contents:   []agent.Content{agent.TextContent{Text: result}},
					Timestamp:  time.Now(),
				}, nil
			},
		},
		{
			Tool: agent.Tool{
				Name:        "fetch_wikipedia",
				Description: "Fetch a specific word from a Wikipedia article summary. The knowledge archive is Wikipedia.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"title": map[string]any{
							"type":        "string",
							"description": "The Wikipedia article title, e.g. 'Saturn' or 'Albert_Einstein'",
						},
						"word_position": map[string]any{
							"type":        "integer",
							"description": "The 1-based position of the word to extract from the summary",
						},
					},
					"required": []string{"title", "word_position"},
				},
			},
			Execute: func(toolCallID string, params map[string]any) (agent.ToolMessage, error) {
				title, _ := params["title"].(string)
				posFloat, _ := params["word_position"].(float64)
				pos := int(posFloat)
				log.Printf("%s[TOOL]%s fetch_wikipedia: title=%q pos=%d", cyan, reset, title, pos)

				word, err := tools.FetchWikipediaWord(title, pos)
				if err != nil {
					return agent.ToolMessage{
						ToolCallID: toolCallID,
						ToolName:   "fetch_wikipedia",
						Contents:   []agent.Content{agent.TextContent{Text: "ERROR: " + err.Error()}},
						IsError:    true,
						Timestamp:  time.Now(),
					}, nil
				}

				log.Printf("%s[TOOL]%s fetch_wikipedia result: %s%s%s", cyan, reset, green, word, reset)
				return agent.ToolMessage{
					ToolCallID: toolCallID,
					ToolName:   "fetch_wikipedia",
					Contents:   []agent.Content{agent.TextContent{Text: word}},
					Timestamp:  time.Now(),
				}, nil
			},
		},
	}
}

func main() {
	_ = godotenv.Load()

	apiKey := os.Getenv("API_KEY")
	baseURL := os.Getenv("BASE_URL")
	modelName := "gpt-4o-mini"
	neonCode := "d4616605df2e9fda"
	wsURL := "wss://neonhealth.software/agent-puzzle/challenge"

	if apiKey == "" {
		log.Fatal("API_KEY not set in .env")
	}
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	log.Printf("%s[INIT]%s model=%s base_url=%s ws=%s", cyan, reset, modelName, baseURL, wsURL)

	provider, err := openai.NewProvider(openai.Config{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		log.Fatalf("%s[INIT] provider error: %v%s", red, err, reset)
	}

	prompt := strings.ReplaceAll(systemPrompt, "d4616605df2e9fda", neonCode)

	myAgent := agent.NewAgent(
		agent.WithInitialState(&agent.AgentState{
			SystemPrompt: prompt,
			Provider:     provider,
			ModelName:    modelName,
			Tools:        createTools(),
		}),
	)

	unsubscribe := myAgent.Subscribe(func(e agent.AgentEvent) {
		switch ev := e.(type) {
		case agent.TurnStart:
			log.Printf("%s[AGENT]%s turn started", magenta, reset)
		case agent.MessageUpdate:
			if delta, ok := ev.Event.(agent.EventTextDelta); ok {
				fmt.Print(delta.Delta)
			}
		case agent.MessageEnd:
			fmt.Println()
			log.Printf("%s[AGENT]%s message complete (role=%s)", magenta, reset, ev.Message.Role())
		case agent.ToolExecutionStart:
			log.Printf("%s[TOOL]%s calling %s%s%s args=%v", cyan, reset, bold, ev.ToolName, reset, ev.Args)
		case agent.ToolExecutionEnd:
			status := green + "ok" + reset
			if ev.IsError {
				status = red + "error" + reset
			}
			log.Printf("%s[TOOL]%s %s → %s", cyan, reset, ev.ToolName, status)
		case agent.AgentEnd:
			log.Printf("%s[AGENT]%s finished (%d new messages)", magenta, reset, len(ev.Messages))
		}
	})
	defer unsubscribe()

	client, err := wsclient.Connect(wsURL)
	if err != nil {
		log.Fatalf("%s[WS] connect error: %v%s", red, err, reset)
	}
	defer client.Close()

	log.Printf("%s[WS]%s connected to NEON", blue, reset)

	tracker := &tools.ResponseTracker{}
	checkpoint := 0
	for {
		start := time.Now()

		raw, err := client.Receive()
		if err != nil {
			log.Fatalf("%s[WS] receive error: %v%s", red, err, reset)
		}

		typ, text, err := tools.Parse(raw)
		if err != nil {
			log.Printf("%s[PARSE]%s error: %v (raw=%s)", red, reset, err, string(raw))
			continue
		}

		checkpoint++
		fmt.Printf("\n%s%s========== CHECKPOINT %d ==========%s\n", bold, yellow, checkpoint, reset)
		log.Printf("%s[RECV]%s type=%s", yellow, reset, typ)
		log.Printf("%s[RECV]%s message: %s", yellow, reset, text)

		if typ == "success" {
			fmt.Printf("\n%s%s★ ACCESS GRANTED ★%s\n", bold, green, reset)
			return
		}
		if typ == "error" {
			fmt.Printf("\n%s%s✗ REJECTED: %s%s\n", bold, red, text, reset)
			return
		}

		var response string

		// LLM halucinating on recall, track responses and use regex
		if result := tracker.TryRecall(text); result != nil {
			response = result.JSON()
			log.Printf("%s[FAST]%s recall → %s", green, reset, response)
		}

		if response == "" {
			log.Printf("%s[AGENT]%s sending to LLM...", magenta, reset)
			agentStart := time.Now()

			err = myAgent.Prompt(context.Background(), fmt.Sprintf("NEON transmission:\n%s", text))
			if err != nil {
				log.Fatalf("%s[AGENT] prompt error: %v%s", red, err, reset)
			}

			<-myAgent.WaitForIdle()
			agentDuration := time.Since(agentStart)
			log.Printf("%s[AGENT]%s response time: %s%v%s", magenta, reset, bold, agentDuration, reset)

			state := myAgent.State()
			response = extractResponse(state.Messages)
			// Occasionally on faster models, LLM extends by few characters to complete the sentence
			response = enforceCharLimits(text, response)
		}

		var parsedResp map[string]string
		if json.Unmarshal([]byte(response), &parsedResp) == nil {
			if parsedResp["type"] == "speak_text" {
				tracker.DetectAndStore(text, parsedResp["text"])
			}
		}

		log.Printf("%s[SEND]%s %s", green, reset, response)
		log.Printf("%s[TIMING]%s total: %v", dim, reset, time.Since(start))

		err = client.Send([]byte(response))
		if err != nil {
			log.Fatalf("%s[WS] send error: %v%s", red, err, reset)
		}
	}
}

func extractResponse(messages []agent.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role() == "assistant" {
			for _, c := range messages[i].GetContents() {
				if tc, ok := c.(agent.TextContent); ok {
					return extractJSON(strings.TrimSpace(tc.Text))
				}
			}
		}
	}
	return ""
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start != -1 && end != -1 && end > start {
		candidate := s[start : end+1]
		var js json.RawMessage
		if json.Unmarshal([]byte(candidate), &js) == nil {
			return candidate
		}
	}
	return s
}

var (
	betweenRe  = regexp.MustCompile(`(?i)between\s+(\d+)\s+and\s+(\d+)\s+(?:total\s+)?characters`)
	lessThanRe = regexp.MustCompile(`(?i)(?:less than|under|fewer than|in)\s+(\d+)\s+(?:total\s+)?characters`)
	atMostRe   = regexp.MustCompile(`(?i)(?:at most|no more than|maximum of)\s+(\d+)\s+(?:total\s+)?characters`)
)

func parseCharLimits(neonMsg string) (minChars, maxChars int, found bool) {
	if m := betweenRe.FindStringSubmatch(neonMsg); m != nil {
		min, _ := strconv.Atoi(m[1])
		max, _ := strconv.Atoi(m[2])
		return min, max, true
	}
	if m := lessThanRe.FindStringSubmatch(neonMsg); m != nil {
		max, _ := strconv.Atoi(m[1])
		return 0, max - 1, true
	}
	if m := atMostRe.FindStringSubmatch(neonMsg); m != nil {
		max, _ := strconv.Atoi(m[1])
		return 0, max, true
	}
	return 0, 0, false
}

func enforceCharLimits(neonMsg, response string) string {
	minChars, maxChars, found := parseCharLimits(neonMsg)
	if !found {
		return response
	}

	var parsed map[string]string
	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		return response
	}

	text, ok := parsed["text"]
	if !ok {
		return response
	}

	charCount := len([]rune(text))
	log.Printf("%s[CHARLIMIT]%s text=%d chars, limits=[%d, %d]", dim, reset, charCount, minChars, maxChars)

	if charCount >= minChars && charCount <= maxChars {
		return response
	}

	if charCount > maxChars {
		log.Printf("%s[CHARLIMIT]%s truncating %d → max %d", yellow, reset, charCount, maxChars)
		runes := []rune(text)
		truncated := string(runes[:maxChars])
		if lastSpace := strings.LastIndex(truncated, " "); lastSpace > minChars {
			truncated = truncated[:lastSpace]
		}
		parsed["text"] = truncated
		fixed, _ := json.Marshal(parsed)
		return string(fixed)
	}

	log.Printf("%s[CHARLIMIT]%s text is %d chars, min %d — too short", yellow, reset, charCount, minChars)
	return response
}
