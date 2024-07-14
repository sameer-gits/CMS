document.addEventListener("DOMContentLoaded", () => {
  const urlParams = new URLSearchParams(window.location.search);
  const email = urlParams.get("email");
  const emailInput = document.getElementById("email") as HTMLInputElement;

  if (email && emailInput) {
    emailInput.value = email;
  }
});
