"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.validate = validate;
exports.validateSync = validateSync;
const child_process_1 = require("child_process");
const path = __importStar(require("path"));
const install_1 = require("./install");
function getBinaryPath() {
    return path.join((0, install_1.getBinaryDir)(), (0, install_1.getBinaryName)());
}
function validate(options = {}) {
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
    return new Promise((resolve, reject) => {
        const proc = (0, child_process_1.spawn)(binaryPath, args, {
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
                    const result = JSON.parse(jsonMatch[0]);
                    resolve(result);
                    return;
                }
                catch {
                    // Fall through to reject
                }
            }
            if (code === 2) {
                reject(new Error(`EnvGuard error: ${stderr || stdout}`));
            }
            else if (code === 1) {
                // Validation failed but we couldn't parse JSON
                reject(new Error(`Validation failed: ${stderr || stdout}`));
            }
            else {
                reject(new Error(`Unexpected exit code ${code}: ${stderr || stdout}`));
            }
        });
        proc.on("error", (err) => {
            reject(new Error(`Failed to run EnvGuard: ${err.message}`));
        });
    });
}
function validateSync(options = {}) {
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
    try {
        const stdout = execFileSync(binaryPath, args, { encoding: "utf-8" });
        return JSON.parse(stdout);
    }
    catch (err) {
        if (err.stdout) {
            try {
                return JSON.parse(err.stdout.toString());
            }
            catch {
                // Fall through
            }
        }
        throw new Error(`EnvGuard failed: ${err.message}`);
    }
}
//# sourceMappingURL=validator.js.map