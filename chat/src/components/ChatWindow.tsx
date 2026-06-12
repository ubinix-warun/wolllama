import { useEffect, useRef, type ReactNode } from "react";
import { MessageBubble } from "./MessageBubble";

export interface ChatMessage {
  role: "user" | "assistant";
  content: string;
}

interface Props {
  messages: ChatMessage[];
  streaming: boolean;
  emptyState?: ReactNode;
}

export function ChatWindow({ messages, streaming, emptyState }: Props) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  if (messages.length === 0 && emptyState) {
    return (
      <div className="flex flex-1 items-center justify-center px-4">
        {emptyState}
      </div>
    );
  }

  return (
    <div className="message-list flex-1 overflow-y-auto px-4 py-6">
      <div className="mx-auto max-w-3xl">
        {messages.map((msg, i) => {
          const isLast = i === messages.length - 1;
          return (
            <MessageBubble
              key={i}
              role={msg.role}
              content={msg.content}
              isStreaming={isLast && streaming && msg.role === "assistant"}
            />
          );
        })}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}
