import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import { execFile } from 'child_process';
import { promisify } from 'util';

const execFileAsync = promisify(execFile);

let diagnosticCollection: vscode.DiagnosticCollection;

export function activate(context: vscode.ExtensionContext) {
  diagnosticCollection = vscode.languages.createDiagnosticCollection('envguard');
  context.subscriptions.push(diagnosticCollection);

  const config = vscode.workspace.getConfiguration('envguard');
  if (!config.get<boolean>('enableValidation', true)) {
    return;
  }

  // Watch .env files for changes
  const envWatcher = vscode.workspace.createFileSystemWatcher('**/.env*');
  context.subscriptions.push(envWatcher);

  envWatcher.onDidChange(validateDocument);
  envWatcher.onDidCreate(validateDocument);
  envWatcher.onDidDelete((uri) => diagnosticCollection.delete(uri));

  // Watch schema file for changes
  const schemaPath = getSchemaPath();
  if (schemaPath) {
    const schemaWatcher = vscode.workspace.createFileSystemWatcher(schemaPath);
    context.subscriptions.push(schemaWatcher);
    schemaWatcher.onDidChange(() => validateAllEnvFiles());
    schemaWatcher.onDidCreate(() => validateAllEnvFiles());
  }

  // Validate all open .env files on activation
  validateAllEnvFiles();

  // Command to manually trigger validation
  const validateCommand = vscode.commands.registerCommand('envguard.validate', async () => {
    const editor = vscode.window.activeTextEditor;
    if (editor) {
      await validateDocument(editor.document.uri);
    }
  });
  context.subscriptions.push(validateCommand);
}

function getSchemaPath(): string | undefined {
  const config = vscode.workspace.getConfiguration('envguard');
  const relativePath = config.get<string>('schemaPath', 'envguard.yaml');
  if (vscode.workspace.workspaceFolders && vscode.workspace.workspaceFolders.length > 0) {
    return path.join(vscode.workspace.workspaceFolders[0].uri.fsPath, relativePath);
  }
  return undefined;
}

async function validateAllEnvFiles() {
  const envFiles = await vscode.workspace.findFiles('**/.env*', '**/node_modules/**');
  for (const uri of envFiles) {
    await validateDocument(uri);
  }
}

async function validateDocument(uri: vscode.Uri) {
  if (!uri.fsPath.match(/\.env([.\w-]*)$/)) {
    return;
  }

  const schemaPath = getSchemaPath();
  if (!schemaPath || !fs.existsSync(schemaPath)) {
    diagnosticCollection.delete(uri);
    return;
  }

  try {
    const diagnostics = await runEnvGuard(uri.fsPath, schemaPath);
    diagnosticCollection.set(uri, diagnostics);
  } catch (err) {
    // EnvGuard binary not found or other error — silently skip
    diagnosticCollection.delete(uri);
  }
}

interface ValidationError {
  key: string;
  message: string;
  rule: string;
}

interface ValidationResult {
  valid: boolean;
  errors: ValidationError[];
  warnings: ValidationError[];
}

async function runEnvGuard(envPath: string, schemaPath: string): Promise<vscode.Diagnostic[]> {
  const envguardPath = await findEnvGuardBinary();
  if (!envguardPath) {
    throw new Error('EnvGuard binary not found');
  }

  const { stdout } = await execFileAsync(envguardPath, [
    'validate',
    '--schema', schemaPath,
    '--env', envPath,
    '--format', 'json'
  ], { timeout: 10000 });

  const result: ValidationResult = JSON.parse(stdout);
  const diagnostics: vscode.Diagnostic[] = [];

  for (const err of result.errors) {
    const range = findKeyRange(envPath, err.key);
    const diagnostic = new vscode.Diagnostic(
      range,
      `${err.key}: ${err.message}`,
      vscode.DiagnosticSeverity.Error
    );
    diagnostic.code = err.rule;
    diagnostic.source = 'envguard';
    diagnostics.push(diagnostic);
  }

  for (const warn of result.warnings) {
    const range = findKeyRange(envPath, warn.key);
    const diagnostic = new vscode.Diagnostic(
      range,
      `${warn.key}: ${warn.message}`,
      vscode.DiagnosticSeverity.Warning
    );
    diagnostic.code = warn.rule;
    diagnostic.source = 'envguard';
    diagnostics.push(diagnostic);
  }

  return diagnostics;
}

async function findEnvGuardBinary(): Promise<string | undefined> {
  // Check PATH
  try {
    const { stdout } = await execFileAsync('which', ['envguard'], { timeout: 5000 });
    const path = stdout.trim();
    if (path) return path;
  } catch {
    // not on PATH
  }

  // Check common locations
  const candidates = [
    path.join(process.env.HOME || '', '.envguard', 'bin', 'envguard'),
    '/usr/local/bin/envguard',
    '/usr/bin/envguard',
  ];

  for (const candidate of candidates) {
    if (fs.existsSync(candidate)) {
      return candidate;
    }
  }

  return undefined;
}

function findKeyRange(filePath: string, key: string): vscode.Range {
  try {
    const content = fs.readFileSync(filePath, 'utf-8');
    const lines = content.split('\n');
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      const idx = line.indexOf(key + '=');
      if (idx !== -1) {
        const start = new vscode.Position(i, idx);
        const end = new vscode.Position(i, line.length);
        return new vscode.Range(start, end);
      }
    }
  } catch {
    // ignore
  }
  return new vscode.Range(0, 0, 0, 0);
}

export function deactivate() {
  diagnosticCollection.dispose();
}
