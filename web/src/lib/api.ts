const API_BASE = '/api'

async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
  authToken?: string
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  }

  if (authToken) {
    headers['Authorization'] = authToken
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  })

  if (!response.ok) {
    const text = await response.text()
    throw new Error(`API error ${response.status}: ${text}`)
  }

  return response.json()
}

export interface CreateInvoiceResponse {
  payment_id: string
  bolt11: string
  payment_hash: string
}

export interface Payment {
  ID: string
  Bolt11: string
  AmountSats: number
  Memo: string
  SenderPubkey: string
  ReceiverPubkey: string
  PaymentHash: string
  Status: string
  CreatedAt: string
  SettledAt: string | null
}

export const api = {
  health: () => apiFetch<{ status: string }>('/health'),

  createInvoice: (amountSats: number, memo: string, token: string) =>
    apiFetch<CreateInvoiceResponse>(
      '/payments/invoice',
      {
        method: 'POST',
        body: JSON.stringify({ amount_sats: amountSats, memo }),
      },
      token
    ),

  getPayment: (id: string, token: string) =>
    apiFetch<Payment>(`/payments/${id}`, {}, token),

  getPaymentHistory: (token: string) =>
    apiFetch<Payment[]>('/payments/history', {}, token),
}
