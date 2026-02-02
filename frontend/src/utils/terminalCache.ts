const cacheKey = (sessionId: string) => `terminal-cache:${sessionId}`;

export const saveTerminalCache = (sessionId: string, buffer: string) => {
  localStorage.setItem(cacheKey(sessionId), buffer);
};

export const loadTerminalCache = (sessionId: string) => {
  return localStorage.getItem(cacheKey(sessionId)) || "";
};
