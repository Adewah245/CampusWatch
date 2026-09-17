const adminLoginForm = document.getElementById("adminLoginForm");
const userLoginForm = document.getElementById("userLoginForm");


// Admin login

if (adminLoginForm) {

    adminLoginForm.addEventListener("submit", function(event) {

        event.preventDefault();

        const email = document.getElementById("adminEmail").value;
        const password = document.getElementById("adminPassword").value;

        if (email === "" || password === "") {

            alert("Please enter your email and password.");

            return;
        }

        alert("Admin login button clicked successfully!");

    });

}


// User login

if (userLoginForm) {

    userLoginForm.addEventListener("submit", function(event) {

        event.preventDefault();

        const email = document.getElementById("userEmail").value;
        const password = document.getElementById("userPassword").value;

        if (email === "" || password === "") {

            alert("Please enter your email and password.");

            return;
        }

        alert("User login button clicked successfully!");

    });

}