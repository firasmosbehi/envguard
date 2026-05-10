export interface ValidationError {
  key: string;
  message: string;
  rule: string;
}

export interface ValidationResult {
  valid: boolean;
  errors: ValidationError[];
  warnings: ValidationError[];
}

export interface ValidateOptions {
  schemaPath?: string;
  envPath?: string;
  strict?: boolean;
}
