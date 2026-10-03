// State: Keep in-memory conversation history to send with each request
let conversationHistory = [];

const chatMessages = document.getElementById("chat-messages");
const chatForm = document.getElementById("chat-form");
const userInput = document.getElementById("user-input");
const sendBtn = document.getElementById("send-btn");
const typingIndicator = document.getElementById("typing-indicator");
const clearBtn = document.getElementById("clear-btn");

// Helper: Scroll chat window to the newest message
function scrollToBottom() {
  chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Helper: Render simple markdown (bold, code, links) for nicer output
function formatText(text) {
  let formatted = text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");

  // Bold **text**
  formatted = formatted.replace(/\*\*(.*?)\*\*/g, "<strong>$1</strong>");
  // Inline `code`
  formatted = formatted.replace(/`([^`]+)`/g, "<code>$1</code>");

  return formatted;
}

// Helper: Append a message bubble to the chat container
function appendMessage(role, text, isError = false) {
  const messageDiv = document.createElement("div");
  messageDiv.classList.add("message");
  messageDiv.classList.add(role === "user" ? "user-message" : "bot-message");
  if (isError) messageDiv.classList.add("error-message");

  const avatarDiv = document.createElement("div");
  avatarDiv.classList.add("avatar");
  avatarDiv.textContent = role === "user" ? "👤" : "🤖";

  const bubbleDiv = document.createElement("div");
  bubbleDiv.classList.add("bubble");
  bubbleDiv.innerHTML = formatText(text);

  messageDiv.appendChild(avatarDiv);
  messageDiv.appendChild(bubbleDiv);
  chatMessages.appendChild(messageDiv);

  scrollToBottom();
}

// Handle Form Submission
chatForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  const message = userInput.value.trim();
  if (!message) return;

  // 1. Add user message to UI
  appendMessage("user", message);
  userInput.value = "";
  userInput.disabled = true;
  sendBtn.disabled = true;

  // 2. Show typing indicator
  typingIndicator.classList.remove("hidden");
  scrollToBottom();

  try {
    // 3. Send message & current conversation history to our Go backend
    const response = await fetch("/api/chat", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        message: message,
        history: conversationHistory,
      }),
    });

    const data = await response.json();

    // 4. Hide typing indicator
    typingIndicator.classList.add("hidden");

    if (!response.ok || data.error) {
      appendMessage("bot", data.error || "An unexpected error occurred.", true);
    } else {
      // 5. Display AI response and update our conversation history state
      appendMessage("bot", data.reply);
      conversationHistory = data.history;
    }
  } catch (err) {
    typingIndicator.classList.add("hidden");
    appendMessage("bot", "Network error: Could not reach Go server.", true);
    console.error(err);
  } finally {
    userInput.disabled = false;
    sendBtn.disabled = false;
    userInput.focus();
  }
});

// Clear Chat button
clearBtn.addEventListener("click", () => {
  conversationHistory = [];
  chatMessages.innerHTML = `
    <div class="message bot-message">
      <div class="avatar">🤖</div>
      <div class="bubble">
        Conversation cleared! How can I help you?
      </div>
    </div>
  `;
  userInput.focus();
});
