const modal = document.getElementById("challenge-modal");

modal.style.display = "flex";
void modal.offsetWidth;
modal.classList.add("show");

const base64 = btoa;

const passwordInput = document.getElementById("required-password-input");

passwordInput.focus();

passwordInput.addEventListener("keypress", async (event) => {
  if (event.key === "Enter") {
    event.preventDefault();

    const slug = window.location.pathname.split("/").pop();
    const plaintextPassword = passwordInput.value;

    const response = await fetch(`/${slug}`, {
      method: "GET",
      headers: { Authorization: `Basic ${base64(plaintextPassword)}` },
    });

    if (response.status === 200) {
      const data = await response.json();

      if (data.url) {
        window.location.href = data.url;
      } else {
        const json = JSON.stringify(data, null, 2);
        const blob = new Blob([json], { type: "application/json" });
        const url = URL.createObjectURL(blob);
        window.open(url, "_blank");
        URL.revokeObjectURL(url);
      }
      return;
    }

    if (response.status === 401) {
      document.querySelector(".invalid-password").style.display = "block";
      document.querySelector(".password-tip").style.display = "none";
      return;
    }
  }
});
