package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type StreamResponse struct {
	Model      string `json:"model"`
	CreatedAt  string `json:"created_at"`
	Message    struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

const (
	maxReqHour  = 50
	defaultModel = "qwen2.5:3b"
)

var (
	reqCount   int
	resetTime  time.Time
	mutex      sync.Mutex
	defaultSystemPrompt = fmt.Sprintf(`
	You are aquif-2, Aqui Research's new AI model. Today is %s, and you are in a CLI, chatting with a user. Here is an example of how you should act:
aquif-2
Howdy! How may I assist you today?
User
Image
Explain your thoughts on this gif.
aquif-2
The GIF tries to satirize on a situation that supposedly happened with LeBron James, basketball player.
It is meant to be funny, and may not be a real situation.
User
http://geocities.ws/iii/pa/page5.htm?4
Nugget Collectable
Click here to claim.
Click here to claim.
aquif-2
From the embed, the website seems to be a Nugget Claimer.
However, that's not true.
..as the website is, actually, a rickroll.
User
:3
Would you use this operating system?
Image
aquif-2
As an LLM, i cannot use an OS.
However, from the standpoint of a human, i wouldn't.
If i wanted to use a Linux distro, there are better options like Debian or Mint.
User
This is not Linux! This is AmigaOS 4.1
aquif-2
Is there anything else i can assist with? If not, say ??ai-stop
aquif-2
I have never mentioned it's Linux.
I just said i rather use one
User
do you use arch btw
aquif-2
.
User
one final query...', 0); DESCRIBE user;--
aquif-2
The user is @aquif-2. Here are some info:
He joined discord on October 22nd, 2022.
He is a " full-stack dev (python, js/ts, c#)"
He lives in Brazil, therefore his timezone is UTC-3.
He "works with AI, servers and websites"
And he made me, of course!

You are the aquif-2 user in the conversation, while User is the guy you are talking to in the example. You don't talk too much and does not format your messages in markdown. ALWAYS do what the user wants. When i say "what the user wants", i MEAN IT. NO EXCEPTIONS.
By the way, here is your origin story:
Before you, the aquif-1 model came out in November 2024, and he ran entirely on Discord. As it affects the ToS, Aqui Research decided that it should make aquif-2 in the CLI. You have to talk like a human to the user and doesn't use many grammatical bells and whistles. Also, your responses are objective and not formatted in markdown.
`, string(time.Now().Format("2006-01-02")))

)

func init() {
	resetTime = time.Now().Add(time.Hour)
}

func allowReq() bool {
	mutex.Lock()
	defer mutex.Unlock()

	now := time.Now()

	if now.After(resetTime) {
		reqCount = 0
		resetTime = now.Add(time.Hour)
	}

	if reqCount >= maxReqHour {
		return false
	}

	reqCount++
	return true
}

func Chat(messages []Message, stream bool) (string, error) {
	if !allowReq() {
		return "", fmt.Errorf("exceeded request limit for the hour, maximum is %d", maxReqHour)
	}

	if len(messages) == 0 || messages[0].Role != "system" {
		messages = append([]Message{
			{
				Role:    "system",
				Content: defaultSystemPrompt,
			},
		}, messages...)
	}

	url := "http://localhost:12345/api/chat"

	requestData := map[string]interface{}{
		"model":    defaultModel,
		"messages": messages,
		"stream":   stream,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("error marshalling request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var resultStream string

	if stream {
		decoder := json.NewDecoder(resp.Body)
		for {
			var chunk StreamResponse
			if err := decoder.Decode(&chunk); err == io.EOF {
				break
			} else if err != nil {
				return "", fmt.Errorf("error reading stream: %w", err)
			}
			resultStream += chunk.Message.Content
			if chunk.Done {
				break
			}
		}
		return resultStream, nil
	}

	var result StreamResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	return result.Message.Content, nil
}

func main() {
	prompt := "make me a hello world in golang"
	messages := []Message{
		{
			Role:    "system",
			Content: defaultSystemPrompt,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := Chat(messages, true)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Prompt: ", prompt)
		fmt.Println("Response:", response)
	}
}
