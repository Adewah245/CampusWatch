const adminLoginForm = document.getElementById("adminLoginForm");

adminLoginForm.addEventListener("submit", function(event) {

    event.preventDefault();

    const email = document.getElementById("adminEmail").value;
    const password = document.getElementById("adminPassword").value;

    if (email === "" || password === "") {
        alert("Please enter your email and password.");
        return;
    }

    alert("Login button clicked successfully!");

});