document.addEventListener('DOMContentLoaded', () => {
    
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


    // Handle expense registration
    const registerExpensesForm = document.getElementById('registerExpensesForm');
    if (registerForm) {
        registerForm.addEventListener('submit', async (event) => {
            event.preventDefault();
            const name = document.getElementById('name').value;
            const description = document.getElementById('description').value;
            const amount = document.getElementById('amount').value;
            const category = document.getElementById('category').value;

            try {
                const response = await fetch('/expenses', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name,description,amount, category })
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

    // Pay expense

    // Handle expense registration
    const payExpense = document.getElementById('payExpense');
    if (registerForm) {
        registerForm.addEventListener('submit', async (event) => {
            event.preventDefault();
            const name = document.getElementById('expenseId').value;
            try {
                const response = await fetch('/expenses/pay', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ expenseId })
                });
                if (response.ok) {
                    alert('Payment successful');
                    registerForm.reset();
                } else {
                    const errorData = await response.json();
                    alert(`Payment failed: ${errorData.message}`);
                }
            } catch (error) {
                alert('An error occurred during registration');
            }
        });
    }
    
});
