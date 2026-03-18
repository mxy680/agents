// Server -> Client
export type ServerMessage =
  | { type: "frame"; data: string } // base64 JPEG
  | { type: "status"; status: SessionStatus }
  | { type: "cookies"; success: boolean; error?: string }
  | { type: "viewport"; width: number; height: number }

export type SessionStatus =
  | "loading"
  | "ready"
  | "login_detected"
  | "extracting"
  | "complete"
  | "error"
  | "timeout"

// Client -> Server
export type ClientMessage =
  | { type: "click"; x: number; y: number }
  | { type: "mousemove"; x: number; y: number }
  | { type: "mousedown"; x: number; y: number }
  | { type: "mouseup"; x: number; y: number }
  | { type: "keydown"; key: string }
  | { type: "keypress"; text: string }
  | { type: "scroll"; deltaX: number; deltaY: number }
  | { type: "ping" }
