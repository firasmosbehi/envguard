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
exports.getBinaryDir = getBinaryDir;
exports.getBinaryName = getBinaryName;
exports.getPlatform = getPlatform;
const fs = __importStar(require("fs"));
const https = __importStar(require("https"));
const os = __importStar(require("os"));
const path = __importStar(require("path"));
const VERSION = "1.0.0";
const REPO = "firasmosbehi/envguard";
function getPlatform() {
    const platform = os.platform();
    const arch = os.arch();
    const platformMap = {
        darwin: "darwin",
        linux: "linux",
        win32: "windows",
    };
    const archMap = {
        x64: "amd64",
        arm64: "arm64",
    };
    const p = platformMap[platform];
    const a = archMap[arch];
    if (!p || !a) {
        throw new Error(`Unsupported platform: ${platform}/${arch}`);
    }
    if (p === "darwin" && a === "amd64" && process.env.APPLE_SILICON_ROSETTA) {
        // Allow override for testing
    }
    return `${p}-${a}`;
}
function getBinaryName() {
    return os.platform() === "win32" ? "envguard.exe" : "envguard";
}
function getBinaryDir() {
    // Store binary inside the package's dist folder
    return path.join(__dirname);
}
function downloadFile(url, dest) {
    return new Promise((resolve, reject) => {
        const file = fs.createWriteStream(dest);
        https
            .get(url, { headers: { "User-Agent": "envguard-node-installer" } }, (response) => {
            if (response.statusCode === 302 || response.statusCode === 301) {
                const redirectUrl = response.headers.location;
                if (!redirectUrl) {
                    reject(new Error("Redirect without location header"));
                    return;
                }
                downloadFile(redirectUrl, dest).then(resolve).catch(reject);
                return;
            }
            if (response.statusCode !== 200) {
                reject(new Error(`Download failed with status ${response.statusCode}`));
                return;
            }
            response.pipe(file);
            file.on("finish", () => {
                file.close();
                resolve();
            });
        })
            .on("error", (err) => {
            fs.unlink(dest, () => { });
            reject(err);
        });
    });
}
async function main() {
    const platform = getPlatform();
    const binaryName = getBinaryName();
    const binaryDir = getBinaryDir();
    const binaryPath = path.join(binaryDir, binaryName);
    // Check if binary already exists
    if (fs.existsSync(binaryPath)) {
        console.log(`EnvGuard binary already exists at ${binaryPath}`);
        return;
    }
    const url = `https://github.com/${REPO}/releases/download/v${VERSION}/envguard-${platform}${os.platform() === "win32" ? ".exe" : ""}`;
    console.log(`Downloading EnvGuard ${VERSION} for ${platform}...`);
    console.log(`URL: ${url}`);
    try {
        await downloadFile(url, binaryPath);
        fs.chmodSync(binaryPath, 0o755);
        console.log(`EnvGuard binary installed at ${binaryPath}`);
    }
    catch (err) {
        console.warn(`Failed to download EnvGuard binary: ${err}`);
        console.warn("You can manually download it from:");
        console.warn(`https://github.com/${REPO}/releases/tag/v${VERSION}`);
        console.warn("The binary will be downloaded on first use.");
        // Non-fatal: binary will be lazy-downloaded on first validate() call
    }
}
// Run if called directly (postinstall)
if (require.main === module) {
    main().catch((err) => {
        console.error(err);
        process.exit(1);
    });
}
//# sourceMappingURL=install.js.map