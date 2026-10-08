import axios from "axios";

export function authErrorCode(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const code = (error.response?.data as { error?: { code?: string } } | undefined)?.error?.code;
    if (code) return code;
  }
  return "UNKNOWN";
}

export function emailCodeSendErrorKey(error: { code?: string; status?: number } | null): string {
  switch (error?.code) {
    case "over_email_send_rate_limit":
    case "over_request_rate_limit":
      return "auth.errors.codeSendRateLimit";
    case "email_address_not_authorized":
    case "otp_disabled":
    case "email_provider_disabled":
    case "signup_disabled":
      return "auth.errors.codeSendUnavailable";
    case "captcha_failed":
      return "auth.errors.codeSendCaptcha";
    case "email_address_invalid":
      return "auth.errors.codeSendInvalidEmail";
    default:
      return error?.status === 429 ? "auth.errors.codeSendRateLimit" : "auth.errors.codeSend";
  }
}
