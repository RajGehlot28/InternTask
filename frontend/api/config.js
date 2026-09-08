const fs = require("fs");
const path = require("path");

// Vercel Serverless Function to expose BACKEND_URL from Vercel Environment Variables or .env
module.exports = (req, res) => {
  res.setHeader("Access-Control-Allow-Origin", "*");
  res.setHeader("Access-Control-Allow-Methods", "GET");
  res.setHeader("Content-Type", "application/json");

  let backendUrl = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || "";

  if (!backendUrl) {
    const candidatePaths = [
      path.join(process.cwd(), ".env"),
      path.join(process.cwd(), "frontend", ".env"),
      path.join(__dirname, "..", ".env"),
      path.join(__dirname, ".env"),
    ];
    for (const p of candidatePaths) {
      try {
        if (fs.existsSync(p)) {
          const content = fs.readFileSync(p, "utf8");
          const match = content.match(/BACKEND_URL\s*=\s*(.+)/);
          if (match) {
            backendUrl = match[1].trim().replace(/^["']|["']$/g, "");
            break;
          }
        }
      } catch (e) {}
    }
  }

  res.status(200).json({ backendUrl });
};
