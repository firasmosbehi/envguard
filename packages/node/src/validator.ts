import { spawn } from "child_process";
import * as path from "path";
import type { ValidateOptions, ValidationResult } from "./types";
import { getBinaryDir, getBinaryName } from "./install";

function getBinaryPath(): string {
  return path.join(getBinaryDir(), getBinaryName());
}

export function validate(options: ValidateOptions = {}): Promise<ValidationResult> {
  const binaryPath = getBinaryPath();
  const args = ["validate", "--format", "json"];

  if (options.schemaPath) {
    args.push("--schema", options.schemaPath);
  }
  if (options.envPath) {
    args.push("--env", options.envPath);
  }
  if (options.strict) {
    args.push("--strict");
  }
  if (options.envName) {
    args.push("--env-name", options.envName);
  }

  return new Promise((resolve, reject) => {
    const proc = spawn(binaryPath, args, {
      stdio: ["ignore", "pipe", "pipe"],
    });

    let stdout = "";
    let stderr = "";

    proc.stdout.on("data", (data) => {
      stdout += data.toString();
    });

    proc.stderr.on("data", (data) => {
      stderr += data.toString();
    });

    proc.on("close", (code) => {
      // Try to parse JSON from stdout regardless of exit code
      const jsonMatch = stdout.match(/\{[\s\S]*\}/);
      if (jsonMatch) {
        try {
          const result = JSON.parse(jsonMatch[0]) as ValidationResult;
          resolve(result);
          return;
        } catch {
          // Fall through to reject
        }
      }

      if (code === 2) {
        reject(new Error(`EnvGuard error: ${stderr || stdout}`));
      } else if (code === 1) {
        // Validation failed but we couldn't parse JSON
        reject(new Error(`Validation failed: ${stderr || stdout}`));
      } else {
        reject(new Error(`Unexpected exit code ${code}: ${stderr || stdout}`));
      }
    });

    proc.on("error", (err) => {
      reject(new Error(`Failed to run EnvGuard: ${err.message}`));
    });
  });
}

export function validateSync(options: ValidateOptions = {}): ValidationResult {
  const { execFileSync } = require("child_process");
  const binaryPath = getBinaryPath();
  const args = ["validate", "--format", "json"];

  if (options.schemaPath) {
    args.push("--schema", options.schemaPath);
  }
  if (options.envPath) {
    args.push("--env", options.envPath);
  }
  if (options.strict) {
    args.push("--strict");
  }
  if (options.envName) {
    args.push("--env-name", options.envName);
  }

  try {
    const stdout = execFileSync(binaryPath, args, { encoding: "utf-8" });
    return JSON.parse(stdout) as ValidationResult;
  } catch (err: any) {
    if (err.stdout) {
      try {
        return JSON.parse(err.stdout.toString()) as ValidationResult;
      } catch {
        // Fall through
      }
    }
    throw new Error(`EnvGuard failed: ${err.message}`);
  }
}
