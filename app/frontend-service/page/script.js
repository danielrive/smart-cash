document.addEventListener('DOMContentLoaded', () => {
    
    // Handle user registration
    const registerForm = document.getElementById('registerForm');
    if (registerForm) {
        registerForm.addEventListener('submit', async (event) => {
            event.preventDefault();
            const firstname = document.getElementById('firstname').value;
            const lastname = document.getElementById('lastname').value;
            const username = document.getElementById('username').value;
            const email = document.getElementById('email').value;
            const password = document.getElementById('password').value;

            try {
                const response = await fetch('http://user.develop.svc.cluster.local:8181/user', {
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

    // Handle user login
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', (event) => {
            event.preventDefault();
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;

            const user = users.find(user => user.username === username && user.password === password);
            if (user) {
                alert('Login successful');
                localStorage.setItem('loggedInUser', username);
                window.location.href = 'expenses.html';
            } else {
                alert('Invalid username or password');
            }
        });
    }

    // Handle expense registration
    const expenseForm = document.getElementById('expenseForm');
    const expensesList = document.getElementById('expensesList');
    if (expenseForm) {
        const displayExpenses = () => {
            expensesList.innerHTML = '';
            expenses.forEach((expense, index) => {
                const expenseItem = document.createElement('li');
                expenseItem.innerHTML = `
                    ${expense.name} - ${expense.date} - ${expense.value} ${expense.currency}
                    <button onclick="removeExpense(${index})">Remove</button>
                `;
                expensesList.appendChild(expenseItem);
            });
        };

        expenseForm.addEventListener('submit', (event) => {
            event.preventDefault();
            const name = document.getElementById('name').value;
            const date = document.getElementById('date').value;
            const value = document.getElementById('value').value;
            const currency = document.getElementById('currency').value;

            const expense = { name, date, value, currency };
            expenses.push(expense);

            localStorage.setItem('expenses', JSON.stringify(expenses));
            displayExpenses();

            expenseForm.reset();
        });

        window.removeExpense = (index) => {
            expenses.splice(index, 1);
            localStorage.setItem('expenses', JSON.stringify(expenses));
            displayExpenses();
        };

        displayExpenses();
    }
});
