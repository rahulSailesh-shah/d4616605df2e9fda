# NEON Authentication Challenge

## Setup

```bash
cp .env.example .env
# Add your API_KEY and BASE_URL
make agent
```

## Reflections

The timeout on every response was both fun and challenging. In a previous project, I had already built an agent harness I could've reused here, but I wanted to push myself and see how far I could get using only local models. Timeouts ended up being one of the biggest challenges.

### Tools Used

- I had previously developed my own agent harness built on top of the OpenAI SDK, which I used as the foundation for this challenge:
  - [go-pi-agent](https://github.com/rahulSailesh-shah/go-pi-agent) — Agent loop with tool calling and streaming
  - [go-pi-ai](https://github.com/rahulSailesh-shah/go-pi-ai) — LLM provider abstraction layer
  - [go-pi-slack-agent](https://github.com/rahulSailesh-shah/go-pi-slack-agent) — Production Slack agent built on the same stack
- I also used Anthropic's Claude Code to speed up boilerplate tasks like setting up the WebSocket client and logging.

### Where the Model Struggled

**Word recall:** Few places where the model hesitated was in recalling a particular word from the context. To overcome that, I tracked the responses separately and then did a regex-based approach on that to extract the exact word being asked.

**Character limits:** The second place is where there's a character limit set on the response. Since I'm using GPT-4o Mini here, which is a fast model, not necessarily a smart model, it sometimes tends to add a couple of characters in order to complete the sentence. I had to enforce a post-processing character limit to solve that.

### Feedback on the Challenge

If the goal of the challenge was to test an AI agent and not a rule-based system, then the text/prompts should've been less predictable. Right now, a lot of it could be solved with regex and pattern matching alone. If the wording, constraints, and message formats changed more randomly, regex-based approaches would break much more often and actual reasoning/adaptability would matter more.
