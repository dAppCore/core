export interface ChatMessage {
  role: 'system' | 'user' | 'assistant';
  content: string;
}

export interface ChatCompletionRequest {
  model: string;
  messages: ChatMessage[];
  max_tokens: number;
  temperature: number;
  stream: boolean;
}

export interface ChatCompletionChunk {
  id: string;
  object: string;
  created: number;
  model: string;
  choices: Array<{
    delta: {
      role?: string;
      content?: string;
    };
    index: number;
    finish_reason: string | null;
  }>;
}

export interface LemSendDetail {
  text: string;
}
