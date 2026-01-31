const defaultModel = 'gemini-2.5-flash-lite';

export type SummaryResult = {
  summaryJa: string;
  language: string;
};

export async function summarizeToJapanese(params: { title: string; text: string }) {
  const apiKey = process.env.GEMINI_API_KEY;
  if (!apiKey) {
    throw new Error('Missing GEMINI_API_KEY');
  }

  const model = process.env.GEMINI_MODEL ?? defaultModel;
  const prompt = buildPrompt(params.title, params.text);
  const url = `https://generativelanguage.googleapis.com/v1beta/models/${model}:generateContent?key=${apiKey}`;

  const response = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      contents: [{
        role: 'user',
        parts: [{ text: prompt }]
      }],
      generationConfig: {
        temperature: 0.2,
        maxOutputTokens: 256
      }
    })
  });

  if (!response.ok) {
    throw new Error(`Gemini API error: ${response.status}`);
  }

  const payload = await response.json();
  const text = payload?.candidates?.[0]?.content?.parts?.map((part: any) => part.text ?? '').join('') ?? '';
  const parsed = safeJsonParse(text);

  if (!parsed?.summary_ja) {
    return fallbackSummary(params.text);
  }

  return {
    summaryJa: String(parsed.summary_ja).trim(),
    language: String(parsed.language ?? 'unknown').trim()
  } as SummaryResult;
}

function buildPrompt(title: string, text: string) {
  const combined = `${title}\n\n${text}`.slice(0, 6000);
  return [
    'You are a technical news summarizer.',
    'Detect the language of the input and output a concise Japanese summary.',
    'Return JSON only in this format:',
    '{"language":"en","summary_ja":"..."}',
    'Guidelines:',
    '- Summary must be 2-4 sentences in Japanese.',
    '- If the input is Japanese, keep the summary in Japanese.',
    '- Do not include markdown or code fences.',
    '',
    `INPUT:\n${combined}`
  ].join('\n');
}

function safeJsonParse(text: string) {
  const start = text.indexOf('{');
  const end = text.lastIndexOf('}');
  if (start === -1 || end === -1 || start >= end) {
    return null;
  }

  try {
    return JSON.parse(text.slice(start, end + 1));
  } catch {
    return null;
  }
}

function fallbackSummary(text: string): SummaryResult {
  const trimmed = text.replace(/\s+/g, ' ').trim().slice(0, 300);
  const isJapanese = /[\u3040-\u30ff\u4e00-\u9faf]/.test(trimmed);

  return {
    summaryJa: trimmed,
    language: isJapanese ? 'ja' : 'unknown'
  };
}
