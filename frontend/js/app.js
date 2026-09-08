// Backend API URL
const API_BASE_URL = "http://localhost:8080";

// Application state
let token = localStorage.getItem("token") || null;
let currentUsername = localStorage.getItem("username") || null;
let currentMode = "login";
let userTickets = [];
let activeFilter = "all";

// DOM elements
const healthBadge = document.getElementById("healthBadge");
const healthText = document.getElementById("healthText");
const toast = document.getElementById("toast");

const authSection = document.getElementById("authSection");
const dashboardSection = document.getElementById("dashboardSection");

const tabLogin = document.getElementById("tabLogin");
const tabRegister = document.getElementById("tabRegister");
const authForm = document.getElementById("authForm");
const formTitle = document.getElementById("formTitle");
const formSubtitle = document.getElementById("formSubtitle");
const usernameInput = document.getElementById("usernameInput");
const passwordInput = document.getElementById("passwordInput");
const authSubmitBtn = document.getElementById("authSubmitBtn");

const loggedUsername = document.getElementById("loggedUsername");
const logoutBtn = document.getElementById("logoutBtn");
const openCreateModalBtn = document.getElementById("openCreateModalBtn");
const createFirstTicketBtn = document.getElementById("createFirstTicketBtn");
const filterBtns = document.querySelectorAll(".filter-btn");
const ticketsGrid = document.getElementById("ticketsGrid");

const ticketModal = document.getElementById("ticketModal");
const closeModalBtn = document.getElementById("closeModalBtn");
const cancelModalBtn = document.getElementById("cancelModalBtn");
const modalOverlay = document.getElementById("modalOverlay");
const createTicketForm = document.getElementById("createTicketForm");
const ticketTitleInput = document.getElementById("ticketTitleInput");
const ticketDescInput = document.getElementById("ticketDescInput");

// Check server health
async function checkHealth() {
  const dot = healthBadge.querySelector(".status-dot");
  try {
    const res = await fetch(`${API_BASE_URL}/health`);
    if (res.ok) {
      const data = await res.json();
      if (data.status === "ok" || data.status === "backend is working") {
        dot.className = "status-dot online";
        healthText.textContent = "Server Online";
        return;
      }
    }
    dot.className = "status-dot offline";
    healthText.textContent = "Server Offline";
  } catch (err) {
    dot.className = "status-dot offline";
    healthText.textContent = "Server Unreachable";
  }
}

// Show toast messages
function showToast(message, type = "success") {
  toast.textContent = message;
  toast.className = `toast ${type}`;
  setTimeout(() => {
    toast.className = "toast hidden";
  }, 4000);
}

// Toggle between Login and Register tabs
function setAuthMode(mode) {
  currentMode = mode;
  if (mode === "login") {
    tabLogin.classList.add("active");
    tabRegister.classList.remove("active");
    formTitle.textContent = "Welcome Back";
    formSubtitle.textContent = "Enter your credentials to access your ticket dashboard";
    authSubmitBtn.textContent = "Login";
  } else {
    tabRegister.classList.add("active");
    tabLogin.classList.remove("active");
    formTitle.textContent = "Create Account";
    formSubtitle.textContent = "Register a new account to manage support tickets";
    authSubmitBtn.textContent = "Register";
  }
}

// Handle login and registration
async function handleAuth(e) {
  e.preventDefault();

  const username = usernameInput.value.trim();
  const password = passwordInput.value.trim();

  if (!username || !password) {
    showToast("Please enter both username and password", "error");
    return;
  }

  const endpoint = currentMode === "login" ? "/auth/login" : "/auth/register";

  try {
    authSubmitBtn.disabled = true;
    authSubmitBtn.textContent = "Processing...";

    const res = await fetch(`${API_BASE_URL}${endpoint}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password })
    });

    const data = await res.json();

    if (!res.ok) {
      throw new Error(data.error || "Authentication failed");
    }

    // Save token and username in local storage
    token = data.token;
    currentUsername = data.username || username;
    localStorage.setItem("token", token);
    localStorage.setItem("username", currentUsername);

    showToast(data.message || "Successfully authenticated!");
    usernameInput.value = "";
    passwordInput.value = "";

    renderAppView();
  } catch (err) {
    showToast(err.message, "error");
  } finally {
    authSubmitBtn.disabled = false;
    authSubmitBtn.textContent = currentMode === "login" ? "Login" : "Register";
  }
}

// Logout user
function logout() {
  token = null;
  currentUsername = null;
  localStorage.removeItem("token");
  localStorage.removeItem("username");
  showToast("Logged out successfully");
  renderAppView();
}

// Fetch user tickets from API
async function fetchTickets() {
  if (!token) return;

  try {
    const res = await fetch(`${API_BASE_URL}/tickets`, {
      headers: {
        "Authorization": `Bearer ${token}`
      }
    });

    if (res.status === 401) {
      logout();
      return;
    }

    if (!res.ok) {
      throw new Error("Failed to load tickets");
    }

    userTickets = await res.json();
    renderTickets();
  } catch (err) {
    showToast(err.message, "error");
  }
}

// Render ticket list
function renderTickets() {
  ticketsGrid.innerHTML = "";

  const filtered = userTickets.filter((ticket) => {
    if (activeFilter === "all") return true;
    return ticket.status === activeFilter;
  });

  if (filtered.length === 0) {
    ticketsGrid.innerHTML = `
      <div class="empty-state">
        <h3>No ${activeFilter !== "all" ? activeFilter.replace("_", " ") : ""} Tickets Found</h3>
        <p>You have no tickets matching this view.</p>
        <button onclick="openModal()" class="primary-btn">+ Create Ticket</button>
      </div>
    `;
    return;
  }

  filtered.forEach((ticket) => {
    const card = document.createElement("div");
    card.className = "ticket-card";

    let actionsHTML = "";
    if (ticket.status === "open") {
      actionsHTML = `
        <button class="action-btn start" onclick="updateStatus(${ticket.id}, 'in_progress')">Start Work</button>
        <button class="action-btn close" onclick="updateStatus(${ticket.id}, 'closed')">Close</button>
      `;
    } else if (ticket.status === "in_progress") {
      actionsHTML = `
        <button class="action-btn close" onclick="updateStatus(${ticket.id}, 'closed')">Close Ticket</button>
      `;
    } else {
      actionsHTML = `<span class="disabled-pill">Closed</span>`;
    }

    const formattedDate = new Date(ticket.created_at || Date.now()).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit"
    });

    card.innerHTML = `
      <div>
        <div class="ticket-header">
          <span class="ticket-id">#TICK-${ticket.id}</span>
          <span class="status-pill ${ticket.status}">${ticket.status.replace("_", " ")}</span>
        </div>
        <h3 class="ticket-title">${escapeHTML(ticket.title)}</h3>
        <p class="ticket-desc">${escapeHTML(ticket.description)}</p>
      </div>
      <div class="ticket-footer">
        <span class="ticket-date">Created: ${formattedDate}</span>
        <div class="ticket-actions">${actionsHTML}</div>
      </div>
    `;

    ticketsGrid.appendChild(card);
  });
}

// Update ticket status
async function updateStatus(ticketId, newStatus) {
  try {
    const res = await fetch(`${API_BASE_URL}/tickets/${ticketId}/status`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      body: JSON.stringify({ status: newStatus })
    });

    const data = await res.json();

    if (!res.ok) {
      throw new Error(data.error || "Status update failed");
    }

    showToast(`Ticket #${ticketId} status updated to ${newStatus.replace("_", " ")}`);
    fetchTickets();
  } catch (err) {
    showToast(err.message, "error");
  }
}

// Create new ticket
async function handleCreateTicket(e) {
  e.preventDefault();

  const title = ticketTitleInput.value.trim();
  const description = ticketDescInput.value.trim();

  if (!title || !description) {
    showToast("Title and Description are required", "error");
    return;
  }

  try {
    const res = await fetch(`${API_BASE_URL}/tickets`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      body: JSON.stringify({ title, description })
    });

    const data = await res.json();

    if (!res.ok) {
      throw new Error(data.error || "Failed to create ticket");
    }

    showToast("Ticket created successfully!");
    closeModal();
    ticketTitleInput.value = "";
    ticketDescInput.value = "";

    fetchTickets();
  } catch (err) {
    showToast(err.message, "error");
  }
}

// Escape HTML helper
function escapeHTML(str) {
  return str.replace(/[&<>'"]/g, (tag) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    "'": "&#39;",
    '"': "&quot;"
  }[tag] || tag));
}

// Modal functions
function openModal() {
  ticketModal.classList.remove("hidden");
}

function closeModal() {
  ticketModal.classList.add("hidden");
}

// Render app view based on login state
function renderAppView() {
  if (token) {
    authSection.classList.add("hidden");
    dashboardSection.classList.remove("hidden");
    loggedUsername.textContent = currentUsername || "User";
    fetchTickets();
  } else {
    authSection.classList.remove("hidden");
    dashboardSection.classList.add("hidden");
  }
}

// Event listeners
tabLogin.addEventListener("click", () => setAuthMode("login"));
tabRegister.addEventListener("click", () => setAuthMode("register"));
authForm.addEventListener("submit", handleAuth);
logoutBtn.addEventListener("click", logout);

openCreateModalBtn.addEventListener("click", openModal);
if (createFirstTicketBtn) createFirstTicketBtn.addEventListener("click", openModal);
closeModalBtn.addEventListener("click", closeModal);
cancelModalBtn.addEventListener("click", closeModal);
modalOverlay.addEventListener("click", closeModal);
createTicketForm.addEventListener("submit", handleCreateTicket);

healthBadge.addEventListener("click", checkHealth);

filterBtns.forEach((btn) => {
  btn.addEventListener("click", (e) => {
    filterBtns.forEach((b) => b.classList.remove("active"));
    e.target.classList.add("active");
    activeFilter = e.target.dataset.filter;
    renderTickets();
  });
});

// Initial run
checkHealth();
renderAppView();
setInterval(checkHealth, 15000);
