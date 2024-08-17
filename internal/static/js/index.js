const vanityUrlToggle = document.getElementById("vanity-url-toggle");
const passwordToggle = document.getElementById("password-toggle");

const workflowInput = document.getElementById("workflow-input");
const vanityUrlInput = document.getElementById("vanity-url-input");
const passwordInput = document.getElementById("password-input");
const shortenButton = document.getElementById("shorten-button");

const workflowErrorMsg = document.getElementById("error-message-workflow");
const vanityUrlErrorMsg = document.getElementById("error-message-vanity-url");
const passwordErrorMsg = document.getElementById("error-message-password");

const ERROR_CODES_TO_MESSAGES = {
  CONTENT_MALFORMED: "Malformed content!",
  SLUG_TAKEN: "Taken! Try another",
  SLUG_RESERVED: "Reserved! Try another",
  SLUG_TOO_SHORT: "Too short! Min 4 chars",
  SLUG_TOO_LONG: "Too long! Max 512 chars",
  SLUG_MISFORMATTED: "Invalid chars! Only base64",
  PASSWORD_TOO_SHORT: "Too short! Min 8 chars",
  PAYLOAD_TOO_LARGE: "Too large! Max 5 MB",
};

function resetForm() {
  vanityUrlToggle.checked = false;
  passwordToggle.checked = false;

  vanityUrlInput.value = "";
  passwordInput.value = "";
  workflowInput.value = "";

  vanityUrlInput.classList.add("hidden");
  passwordInput.classList.add("hidden");

  workflowErrorMsg.style.visibility = "hidden";
  vanityUrlErrorMsg.style.visibility = "hidden";
  passwordErrorMsg.style.visibility = "hidden";

  updateShortenButtonState();
}

document.addEventListener("DOMContentLoaded", () => {
  resetForm();

  workflowInput.addEventListener("input", () => {
    workflowErrorMsg.style.visibility = "hidden";
    updateShortenButtonState();
  });

  vanityUrlInput.addEventListener("input", () => {
    vanityUrlErrorMsg.style.visibility = "hidden";
  });

  passwordInput.addEventListener("input", () => {
    passwordErrorMsg.style.visibility = "hidden";
  });
});

function updateShortenButtonState() {
  const isDisabled = workflowInput.value.trim() === "";

  shortenButton.disabled = isDisabled;

  if (isDisabled) {
    shortenButton.setAttribute(
      "title",
      "Please enter a workflow or URL to shorten"
    );
  } else {
    shortenButton.removeAttribute("title");
  }
}

function toggleVanityUrl() {
  vanityUrlInput.classList.toggle("hidden");
  if (vanityUrlInput.classList.contains("hidden")) {
    vanityUrlInput.value = "";
    vanityUrlErrorMsg.textContent = "";
  }
}

function togglePassword() {
  passwordInput.classList.toggle("hidden");
  if (passwordInput.classList.contains("hidden")) {
    passwordInput.value = "";
    passwordErrorMsg.textContent = "";
  }
}

function throwConfetti() {
  confetti({
    particleCount: 400,
    spread: 90,
    origin: { y: 0.6 },
    zIndex: 2,
  });

  const duration = 0.9 * 1000;
  const end = Date.now() + duration;

  (function frame() {
    confetti({
      particleCount: 7,
      angle: 60,
      spread: 55,
      origin: { x: 0 },
      zIndex: 2,
    });
    confetti({
      particleCount: 7,
      angle: 120,
      spread: 55,
      origin: { x: 1 },
      zIndex: 2,
    });

    if (Date.now() < end) requestAnimationFrame(frame);
  })();
}

async function handleSubmit(event) {
  event.preventDefault();

  const form = event.target;
  const formData = new FormData(form);

  const jsonData = {};
  formData.forEach((value, key) => {
    if (value !== "") jsonData[key] = value;
  });

  let kind = "";

  if (isUrl(jsonData.content)) {
    kind = "url";
  } else if (isJson(jsonData.content)) {
    kind = "workflow";
  } else {
    workflowErrorMsg.textContent = "Malformed content!";
    workflowErrorMsg.style.visibility = "visible";
    return;
  }

  if (jsonData.password?.length < 8) {
    passwordErrorMsg.textContent = "Too short! Min 8 chars";
    passwordErrorMsg.style.visibility = "visible";
    return;
  }

  const response = await fetch(form.action, {
    method: form.method,
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(jsonData),
  });

  const jsonResponse = await response.json();

  if (response.ok) {
    // console.log("Success:", jsonResponse); // @TODO: Remove
    showModal(jsonResponse.data.slug, kind, jsonData.password !== undefined);
    throwConfetti();
    return;
  }

  // console.error("Error:", jsonResponse.error); // @TODO: Remove

  const errorCode = jsonResponse.error.code;
  const errorMsg = ERROR_CODES_TO_MESSAGES[errorCode];

  if (errorCode.startsWith("SLUG") && errorMsg !== undefined) {
    vanityUrlErrorMsg.textContent = errorMsg;
    vanityUrlErrorMsg.style.visibility = "visible";
    return;
  }

  if (
    (errorCode.startsWith("CONTENT") || errorCode.startsWith("PAYLOAD")) &&
    errorMsg !== undefined
  ) {
    workflowErrorMsg.textContent = errorMsg;
    workflowErrorMsg.style.visibility = "visible";
    return;
  }
}

const isUrl = (str) => {
  try {
    new URL(str);
    return true;
  } catch (_) {
    return false;
  }
};

const isJson = (str) => {
  try {
    JSON.parse(str);
    return true;
  } catch (e) {
    return false;
  }
};

function showModal(shortlink, kind, hasPassword = false) {
  const modal = document.getElementById("success-modal");
  const shortlinkText = document.getElementById("shortlinkText");
  const successMessage = document.getElementById("success-message");
  const successTip = document.getElementById("success-tip");

  if (kind === "url" && !hasPassword) {
    successMessage.textContent = "Your shortlink has been created.";
    successTip.innerHTML = "This will permanently redirect to your URL.";
  } else if (kind === "workflow" && !hasPassword) {
    successMessage.textContent =
      "This shortlink will serve your workflow JSON.";
    successTip.innerHTML = "Append <code>/view</code> to display on canvas.";
  } else if (kind === "url" && hasPassword) {
    successMessage.textContent =
      "Your password-protected shortlink has been created.";
    successTip.innerHTML = "Visiting this URL will require your password.";
  } else if (kind === "workflow" && hasPassword) {
    successMessage.textContent =
      "Your password-protected shortlink has been created.";
    successTip.innerHTML =
      "Accessing this workflow will require your password.";
  }

  shortlinkText.textContent = `https://n8n.to/${shortlink}`;

  shortlinkText.addEventListener("click", copyToClipboard);

  modal.style.display = "flex";
  // Trigger reflow
  void modal.offsetWidth;
  modal.classList.add("show");

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") closeModal();
  });
}

function copyToClipboard(event) {
  const shortlink = event.target.textContent;
  navigator.clipboard
    .writeText(shortlink)
    .then(() => {
      const originalText = event.target.textContent;
      const originalBgColor = event.target.style.backgroundColor;

      event.target.textContent = "Copied to clipboard!";
      event.target.style.backgroundColor = "#10b981";

      setTimeout(() => {
        event.target.textContent = originalText;
        event.target.style.backgroundColor = originalBgColor;
      }, 1500);
    })
    .catch((err) => {
      console.error("Failed to copy text: ", err);
    });
}

function closeModal() {
  const modal = document.getElementById("success-modal");
  modal.classList.remove("show");
  setTimeout(() => {
    modal.style.display = "none";
  }, 300);

  resetForm();
}

document
  .getElementById("success-modal")
  .addEventListener("click", function (event) {
    if (event.target === this) closeModal();
  });
