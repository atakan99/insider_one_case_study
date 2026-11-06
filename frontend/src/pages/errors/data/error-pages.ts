interface ErrorPage {
  code: number
  subtitle: string
  error: string
}

type ErrorCode = '404' | '500'
type ErrorPageContent = Omit<ErrorPage, 'code'>

const FALLBACK_ERROR_CODE: ErrorCode = '404'

const ERROR_PAGE_MAP: Record<ErrorCode, ErrorPageContent> = {
  404: {
    subtitle: 'Page Not Found',
    error: 'The page you are looking for might have been removed, had its name changed, or is temporarily unavailable.',
  },
  500: {
    subtitle: 'Internal Server Error',
    error: 'The server encountered an unexpected condition which prevented it from fulfilling the request.',
  },
}

function isErrorCode(code: string): code is ErrorCode {
  return Object.hasOwn(ERROR_PAGE_MAP, code)
}

export function resolveErrorPage(code: string): ErrorPage {
  const resolvedCode = isErrorCode(code) ? code : FALLBACK_ERROR_CODE

  return {
    code: Number(resolvedCode),
    ...ERROR_PAGE_MAP[resolvedCode],
  }
}
