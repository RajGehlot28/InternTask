// Vercel Serverless Function to expose BACKEND_URL from Vercel Environment Variables
module.exports = (req, res) => {
  res.setHeader("Access-Control-Allow-Origin", "*");
  res.setHeader("Access-Control-Allow-Methods", "GET");
  const backendUrl = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || "";
  res.status(200).json({ backendUrl });
};
