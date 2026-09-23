/* =========================================
   ADMIN LOGIN
   ========================================= */

const adminLoginForm = document.getElementById("adminLoginForm");

if (adminLoginForm) {
    adminLoginForm.addEventListener("submit", function (event) {
        event.preventDefault();

        const emailInput = document.getElementById("adminEmail");
        const passwordInput = document.getElementById("adminPassword");

        const email = emailInput ? emailInput.value.trim() : "";
        const password = passwordInput ? passwordInput.value : "";

        if (email === "" || password === "") {
            alert("Please enter your email and password.");
            return;
        }

        alert("Admin login successful!");

        window.location.href = "admin-dashboard.html";
    });
}


/* =========================================
   USER LOGIN
   ========================================= */

const userLoginForm = document.getElementById("userLoginForm");

if (userLoginForm) {

    userLoginForm.addEventListener("submit", function (event) {

        event.preventDefault();


        const emailInput = document.getElementById("userEmail");
        const passwordInput = document.getElementById("userPassword");


        const email = emailInput.value.trim().toLowerCase();
        const password = passwordInput.value;


        /* Check empty fields */

        if (email === "" || password === "") {

            alert("Please enter your email and password.");

            return;
        }


        /* Check password length */

        if (password.length < 8) {

            alert("Password must be at least 8 characters long.");

            return;
        }


        /* Get registered account */

        const savedAccount =
            localStorage.getItem("campusWatchAccount");


        if (!savedAccount) {

            alert(
                "No CampusWatch account was found. " +
                "Please create an account first."
            );

            return;
        }


        const account = JSON.parse(savedAccount);


        /* Check email */

        if (email !== account.email) {

            alert(
                "No account was found with this email address."
            );

            return;
        }


        /* Check password */

        if (password !== account.password) {

            alert("Incorrect password. Please try again.");

            return;
        }


        /*
         * Save the logged-in user's name.
         * The dashboard can use this later.
         */

        localStorage.setItem(
            "campusWatchCurrentUser",
            account.fullName
        );


        /* Welcome message */

        alert(
            "Welcome back to CampusWatch, " +
            account.fullName +
            "!"
        );


        /* Open dashboard */

        window.location.href = "dashboard.html";

    });

}

/* =========================================
   USER REGISTRATION
   ========================================= */

const registerForm = document.getElementById("registerForm");

if (registerForm) {

    registerForm.addEventListener("submit", function (event) {

        event.preventDefault();

        const fullName = document.getElementById("fullName").value.trim();
        const email = document.getElementById("email").value.trim();
        const password = document.getElementById("password").value;
        const confirmPassword = document.getElementById("confirmPassword").value;
        const terms = document.getElementById("terms").checked;


        /* Check required fields */

        if (
            fullName === "" ||
            email === "" ||
            password === "" ||
            confirmPassword === ""
        ) {
            alert("Please fill in all the required fields.");
            return;
        }


        /* Check password length */

        if (password.length < 8) {
            alert("Password must be at least 8 characters long.");
            return;
        }


        /* Check password confirmation */

        if (password !== confirmPassword) {
            alert("Passwords do not match.");
            return;
        }


        /* Check terms */

        if (!terms) {
            alert("Please agree to the CampusWatch terms and conditions.");
            return;
        }


        /*
         * Save the registered account for this
         * frontend demonstration.
         */

        const account = {
            fullName: fullName,
            email: email.toLowerCase(),
            password: password,
        };

        localStorage.setItem(
            "campusWatchAccount",
            JSON.stringify(account)
        );


        /* Welcome message */

        alert(
            "Welcome to CampusWatch, " +
            fullName +
            "! Your account has been created successfully."
        );


        /* Take the user to login */

        window.location.href = "login.html";

    });

}
/* ===================================
   SYSTEMS PAGES
   =================================== */

const systemsTableBody = document.getElementById("systemsTableBody");
const clusterFilter = document.getElementById("clusterFilter");
const statusFilter = document.getElementById("statusFilter");
const systemSearch = document.getElementById("systemSearch");

if (systemsTableBody) {

    const systems = [
        {
            id: "SYS-000001",
            hostname: "LAB-A-PC-001",
            cluster: "Cluster 1",
            table: "Table 1",
            status: "Online",
            user: "John Doe",
            lastActivity: "2 minutes ago"
        },

        {
            id: "SYS-000002",
            hostname: "LAB-A-PC-002",
            cluster: "Cluster 1",
            table: "Table 1",
            status: "Idle",
            user: "Jane Smith",
            lastActivity: "8 minutes ago"
        },

        {
            id: "SYS-000003",
            hostname: "LAB-A-PC-003",
            cluster: "Cluster 1",
            table: "Table 1",
            status: "Offline",
            user: "No user",
            lastActivity: "25 minutes ago"
        },

        {
            id: "SYS-000004",
            hostname: "LAB-A-PC-004",
            cluster: "Cluster 1",
            table: "Table 1",
            status: "Online",
            user: "David James",
            lastActivity: "1 minute ago"
        },

        {
            id: "SYS-000005",
            hostname: "LAB-A-PC-005",
            cluster: "Cluster 1",
            table: "Table 2",
            status: "Maintenance",
            user: "No user",
            lastActivity: "1 hour ago"
        },

        {
            id: "SYS-000006",
            hostname: "LAB-A-PC-006",
            cluster: "Cluster 1",
            table: "Table 2",
            status: "Online",
            user: "Michael",
            lastActivity: "3 minutes ago"
        }
    ];


    function displaySystems(systemList) {

        systemsTableBody.innerHTML = "";

        systemList.forEach(function(system) {

            const row = document.createElement("tr");

            row.innerHTML = `
                <td>${system.id}</td>
                <td>${system.hostname}</td>
                <td>${system.cluster}</td>
                <td>${system.table}</td>
                <td>
                    <span class="table-status ${system.status.toLowerCase()}">
                        ${system.status}
                    </span>
                </td>
                <td>${system.user}</td>
                <td>${system.lastActivity}</td>
                <td>
                    <a href="system-details.html" class="view-button">
                        View
                    </a>
                </td>
            `;

            systemsTableBody.appendChild(row);
        });
    }


    function filterSystems() {

        const selectedCluster = clusterFilter.value;
        const selectedStatus = statusFilter.value;
        const searchText = systemSearch.value.toLowerCase();

        const filteredSystems = systems.filter(function(system) {

            const matchesCluster =
                selectedCluster === "all" ||
                system.cluster === selectedCluster;

            const matchesStatus =
                selectedStatus === "all" ||
                system.status.toLowerCase() === selectedStatus;

            const matchesSearch =
                system.id.toLowerCase().includes(searchText) ||
                system.hostname.toLowerCase().includes(searchText);

            return matchesCluster && matchesStatus && matchesSearch;
        });

        displaySystems(filteredSystems);
    }


    displaySystems(systems);


    clusterFilter.addEventListener("change", filterSystems);
    statusFilter.addEventListener("change", filterSystems);
    systemSearch.addEventListener("input", filterSystems);

}

