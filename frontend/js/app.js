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

        alert("Admin login successful!");

        window.location.href = "admin-dashboard.html"
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

        alert("Login successful! Welcome to CampusWatch.");

        window.location.href = "dashboard.html";
    });
}

// User registration

const registerForm = document.getElementById("registerForm");

if (registerForm) {

    registerForm.addEventListener("submit", function(event) {

        event.preventDefault();

        const fullName = document.getElementById("fullName").value;
        const email = document.getElementById("email").value;
        const password = document.getElementById("password").value;
        const confirmPassword = document.getElementById("confirmPassword").value;
        const terms = document.getElementById("terms").checked;

        if (fullName === "" || email === "" || password === "" || confirmPassword === "") {

            alert("Please fill in all the required fields.");

            return;
        }

        if (password !== confirmPassword) {

            alert("Passwords do not match.");

            return;
        }

        if (!terms) {

            alert("Please agree to the terms and conditions.");

            return;
        }

        alert("Registration form submitted successfully!");

    });

}
// Systems page

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