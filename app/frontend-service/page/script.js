
document.addEventListener('DOMContentLoaded', () => {
    
    // Handle user registration
    const login = document.getElementById('login');
    if (login) {
        login.addEventListener('submit', async (event) => {
            event.preventDefault();
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            try {
                const response = await fetch('/user/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username,password })
                });
                if (response.ok) {
                    alert('Login successful');
                    login.reset();
                } else {
                    const errorData = await response.json();
                    alert(`Login failed: ${errorData.message}`);
                }
            } catch (error) {
                alert('An error occurred during Login');
            }
        });
    }


    // Handle user registration
    const registerForm = document.getElementById('registerUserForm');
    if (registerForm) {
        registerForm.addEventListener('submit', async (event) => {
            event.preventDefault();
            const firstname = document.getElementById('firstname').value;
            const lastname = document.getElementById('lastname').value;
            const username = document.getElementById('username').value;
            const email = document.getElementById('email').value;
            const password = document.getElementById('password').value;

            try {
                const response = await fetch('/user', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ firstname,lastname,username, email,password })
                });
                if (response.ok) {
                    alert('Registration successful');
                    registerForm.reset();
                } else {
                    const errorData = await response.json();
                    alert(`Registration failed: ${errorData.message}`);
                }
            } catch (error) {
                alert('An error occurred during registration');
            }
        });
    }
   
    // list expenses

    async function fetchExpense() {
        // Get the expense ID from the form input
        const expenseId = document.getElementById("expenseId").value;

        // Replace 'YOUR_API_ENDPOINT' with the actual endpoint URL
        const apiUrl = `/expenses/${expenseId}`;

        // Clear any previous results
        document.getElementById("expenseResult").innerHTML = "Loading...";

        try {
            // Make the GET request to the API
            const response = await fetch(apiUrl, {
                method: "GET",
                headers: {
                    "Content-Type": "application/json"
                }
            });

            // Check if the response is successful
            if (!response.ok) {
                throw new Error(`Error: ${response.status} ${response.statusText}`);
            }

            // Parse the JSON response
            const expenseData = await response.json();

            // Display the expense data
            document.getElementById("expenseResult").innerHTML = `
                <h3>Expense Details</h3>
                <p><strong>ID:</strong> ${expenseData.expenseId}</p>
                <p><strong>Amount:</strong> ${expenseData.amount}</p>
                <p><strong>Description:</strong> ${expenseData.description}</p>
                <p><strong>Date:</strong> ${expenseData.date}</p>
            `;
        } catch (error) {
            // Display an error message if something goes wrong
            document.getElementById("expenseResult").innerHTML = `<p style="color: red;">${error.message}</p>`;
        }
    }

    // Attach the fetchExpense function to the button click
    document.getElementById("expenseForm").addEventListener("submit", (event) => {
        event.preventDefault();  // Prevent the form from submitting traditionally
        fetchExpense();          // Call the function to fetch expense
    });



    // Pay expense

    // Handle expense registration
    const payExpense = document.getElementById('payExpense');
    if (payExpense) {
        payExpense.addEventListener('submit', async (event) => {
            event.preventDefault();
            const expenseId = document.getElementById('expenseId').value;
            try {
                const response = await fetch('/expenses/pay', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ expenseId })
                });
                if (response.ok) {
                    alert('Payment successful');
                    payExpense.reset();
                } else {
                    const errorData = await response.json();
                    alert(`Payment failed: ${errorData.message}`);
                }
            } catch (error) {
                alert('An error occurred during registration');
            }
        });
    }

    // List expenses 

    // Handle expense registration
   
});

