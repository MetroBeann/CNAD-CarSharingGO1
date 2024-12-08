// Supabase client initialization
const SUPABASE_URL = 'https://wjdhhzmaclmsvaiszagk.supabase.co';
const SUPABASE_KEY = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6IndqZGhoem1hY2xtc3ZhaXN6YWdrIiwicm9sZSI6ImFub24iLCJpYXQiOjE3MzA1NDkwMDYsImV4cCI6MjA0NjEyNTAwNn0.Vq2hU8ViVOfe7TZRHf1RWTDyfKwwvXNvlLRucnq_hJo';

// Global variables
let selectedVehicle = null;
let vehicles = [];

// Utility function for rental cost calculation
function calculateRentalCost(startTime, endTime, hourlyRate = 9) {
    const start = new Date(startTime);
    const end = new Date(endTime);
    const hours = Math.ceil((end - start) / (1000 * 60 * 60)); // Round up to nearest hour
    return (hours * hourlyRate).toFixed(2);
}

// Initialize page
document.addEventListener('DOMContentLoaded', () => {
    // Initialize state
    updateAuthUI();
    
    // Add event listeners for auth buttons
    document.querySelectorAll('[data-form]').forEach(button => {
        button.addEventListener('click', (e) => {
            // Remove active class from all buttons
            document.querySelectorAll('[data-form]').forEach(btn => 
                btn.classList.remove('active'));
            // Add active class to clicked button
            e.target.classList.add('active');
            showForm(e.target.dataset.form);
        });
    });

    document.getElementById('logoutButton').addEventListener('click', handleLogout);

    // Add event listener for search form
    document.getElementById('searchForm')?.addEventListener('submit', async (e) => {
        e.preventDefault();
        await loadAvailableVehicles();
    });

    // Set default date/time values for search
    const now = new Date();
    const later = new Date(now.getTime() + 2 * 60 * 60 * 1000); // 2 hours later
    
    const startTimeInput = document.getElementById('startTime');
    const endTimeInput = document.getElementById('endTime');
    
    if (startTimeInput && endTimeInput) {
        startTimeInput.value = now.toISOString().slice(0, 16);
        endTimeInput.value = later.toISOString().slice(0, 16);
    }

    // Initially show login form and set active state
    const loginButton = document.querySelector('[data-form="login"]');
    if (loginButton) {
        loginButton.classList.add('active');
        showForm('login');
    }
});

// Authentication Functions
function updateAuthUI() {
    const authToken = localStorage.getItem('authToken');
    const userEmail = localStorage.getItem('userEmail');
    const authSection = document.getElementById('authSection');
    const mainContent = document.getElementById('mainContent');
    const authStatus = document.getElementById('authStatus');
    const logoutBtn = document.getElementById('logoutButton');
    const formSwitchBtns = document.querySelectorAll('.form-switch button');
    const userInfo = document.getElementById('userInfo');

    if (authToken && userEmail) {
        // User is logged in
        authSection.classList.add('hidden');
        mainContent.classList.remove('hidden');
        logoutBtn.classList.remove('hidden');
        formSwitchBtns.forEach(btn => btn.classList.add('hidden'));
        if (userInfo) userInfo.textContent = `Logged in as ${userEmail}`;

        // Initialize vehicle service functionality
        loadAvailableVehicles();
        loadMyBookings();
    } else {
        // User is not logged in
        authSection.classList.remove('hidden');
        mainContent.classList.add('hidden');
        logoutBtn.classList.add('hidden');
        formSwitchBtns.forEach(btn => btn.classList.remove('hidden'));
        if (authStatus) authStatus.textContent = 'Please log in to continue';
    }
}

function showForm(formType) {
    const container = document.getElementById('formContainer');
    const title = formType.charAt(0).toUpperCase() + formType.slice(1);
    
    container.innerHTML = `
        <h2 class="text-2xl font-bold mb-6">${title}</h2>
        <form id="userForm" class="space-y-4">
            <div>
                <input type="email" id="email" class="input-field" 
                       placeholder="Email" required>
            </div>
            <div>
                <input type="password" id="password" class="input-field" 
                       placeholder="Password" required>
            </div>
            ${formType === 'register' ? `
                <div>
                    <input type="tel" id="phoneNumber" class="input-field" 
                           placeholder="Phone Number">
                </div>
                <div>
                    <select id="membershipTier" class="input-field">
                        <option value="">Select Membership Tier</option>
                        <option value="Basic">Basic</option>
                        <option value="Premium">Premium</option>
                        <option value="VIP">VIP</option>
                    </select>
                </div>
            ` : ''}
            <button type="submit" class="nav-button ${formType}">
                ${formType === 'login' ? 'Sign In' : 'Create Account'}
            </button>
        </form>
    `;

    container.classList.remove('hidden');

    document.getElementById('userForm').addEventListener('submit', (e) => {
        e.preventDefault();
        handleSubmit(formType);
    });
}

async function handleSubmit(formType) {
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const phoneNumber = document.getElementById('phoneNumber')?.value;
    const membershipTier = document.getElementById('membershipTier')?.value;

    let endpoint = '';
    let body = {};

    switch (formType) {
        case 'register':
            endpoint = '/users/register';
            body = { email, password, phone_number: phoneNumber, membership_tier: membershipTier };
            break;
        case 'login':
            endpoint = '/users/login';
            body = { email, password };
            break;
    }

    try {
        const response = await fetch(`http://localhost:8080${endpoint}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(body)
        });

        const data = await response.json();

        if (response.ok) {
            if (formType === 'login') {
                localStorage.setItem('authToken', data.token);
                localStorage.setItem('userId', data.user_id);
                localStorage.setItem('userEmail', data.email);
                showMessage('Login successful!', true);
                updateAuthUI();
            } else {
                showMessage('Registration successful! Please log in.', true);
                showForm('login');
                // Set login button as active after registration
                document.querySelectorAll('[data-form]').forEach(btn => 
                    btn.classList.remove('active'));
                document.querySelector('[data-form="login"]').classList.add('active');
            }
        } else {
            showMessage(data.message || 'An error occurred', false);
        }
    } catch (error) {
        showMessage('Failed to process request', false);
    }
}

function handleLogout() {
    localStorage.removeItem('authToken');
    localStorage.removeItem('userId');
    localStorage.removeItem('userEmail');
    updateAuthUI();
    showForm('login');
    showMessage('Logged out successfully', true);
}

function showMessage(message, isSuccess) {
    const messageDiv = document.getElementById('responseMessage');
    messageDiv.className = isSuccess ? 'success-message' : 'error-message';
    messageDiv.textContent = message;
    messageDiv.classList.remove('hidden');

    setTimeout(() => {
        messageDiv.classList.add('hidden');
    }, 3000);
}

// Vehicle Service Functions
async function loadAvailableVehicles() {
    try {
        const response = await fetch(`${SUPABASE_URL}/rest/v1/vehicles?status=eq.available`, {
            headers: {
                'apikey': SUPABASE_KEY,
                'Authorization': `Bearer ${SUPABASE_KEY}`
            }
        });

        if (!response.ok) throw new Error('Failed to fetch vehicles');
        vehicles = await response.json();
        displayVehicles(vehicles);
    } catch (error) {
        console.error('Error loading vehicles:', error);
        showMessage('Failed to load available vehicles', false);
    }
}

function displayVehicles(vehicles) {
    const vehiclesList = document.getElementById('vehiclesList');
    if (!vehiclesList) return;

    vehiclesList.innerHTML = vehicles.map(vehicle => `
        <div class="vehicle-card border rounded-lg p-4 hover:shadow-lg transition-shadow">
            <div class="flex justify-between items-start mb-4">
                <div>
                    <h3 class="font-semibold text-lg">${vehicle.model}</h3>
                    <p class="text-sm text-gray-600">${vehicle.type}</p>
                    <p class="text-sm text-blue-600">$${vehicle.hourly_rate || '9'}/hour</p>
                </div>
                <div class="flex items-center space-x-1">
                    <svg class="w-5 h-5 ${vehicle.battery_level > 50 ? 'text-green-500' : 'text-orange-500'}" 
                         fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                              d="M3 10h14a2 2 0 0 1 2 2v4a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2v-4a2 2 0 0 1 2-2m18 1v6m-3-3h3"/>
                    </svg>
                    <span class="text-sm">${vehicle.battery_level}%</span>
                </div>
            </div>
            
            <div class="space-y-2 mb-4">
                <div class="flex items-center space-x-2">
                    <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                              d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"/>
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                              d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"/>
                    </svg>
                    <span class="text-sm">${vehicle.location}</span>
                </div>
                <div class="flex items-center space-x-2">
                    <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                              d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z"/>
                    </svg>
                    <span class="text-sm capitalize">${vehicle.cleanliness_status.replace('_', ' ')}</span>
                </div>
            </div>

            <div class="flex space-x-2">
                <button onclick="showVehicleDetails(${vehicle.id})" 
                        class="flex-1 px-3 py-2 text-sm bg-gray-100 hover:bg-gray-200 rounded-md">
                    More Info
                </button>
                <button onclick="showReservationConfirmation(${vehicle.id})" 
                        class="flex-1 px-3 py-2 text-sm bg-blue-500 text-white hover:bg-blue-600 rounded-md">
                    Reserve
                </button>
            </div>
        </div>
    `).join('');
}

async function loadMyBookings() {
    try {
        const userId = localStorage.getItem('userId');
        if (!userId) return;

        const response = await fetch(
            `${SUPABASE_URL}/rest/v1/bookings?user_id=eq.${userId}&select=*,vehicle:vehicles(*)`,
            {
                headers: {
                    'apikey': SUPABASE_KEY,
                    'Authorization': `Bearer ${SUPABASE_KEY}`
                }
            }
        );

        if (!response.ok) throw new Error('Failed to fetch bookings');
        const bookings = await response.json();
        displayBookings(bookings);
    } catch (error) {
        console.error('Error loading bookings:', error);
        showMessage('Failed to load bookings', false);
    }
}

function displayBookings(bookings) {
    const bookingsList = document.getElementById('bookingsList');
    if (!bookingsList) return;

    if (!bookings.length) {
        bookingsList.innerHTML = '<p class="text-gray-500">No bookings found.</p>';
        return;
    }

    bookingsList.innerHTML = bookings.map(booking => {
        // Calculate total cost if not present in database
        const totalCost = booking.total_cost || calculateRentalCost(
            booking.start_time, 
            booking.end_time, 
            booking.vehicle.hourly_rate || 9
        );

        return `
            <div class="border rounded-lg p-4">
                <div class="flex justify-between items-start">
                    <div>
                        <h3 class="font-semibold">${booking.vehicle.model}</h3>
                        <p class="text-sm text-gray-600">
                            Start: ${new Date(booking.start_time).toLocaleString()}
                        </p>
                        <p class="text-sm text-gray-600">
                            End: ${new Date(booking.end_time).toLocaleString()}
                        </p>
                        <p class="text-sm font-semibold text-blue-600 mt-2">
                            Total Cost: $${parseFloat(totalCost).toFixed(2)}
                        </p>
                    </div>
                    <span class="px-2 py-1 text-sm rounded-full ${
                        booking.status === 'confirmed' 
                            ? 'bg-green-100 text-green-800' 
                            : 'bg-yellow-100 text-yellow-800'
                    }">
                        ${booking.status.charAt(0).toUpperCase() + booking.status.slice(1)}
                    </span>
                </div>
                <div class="flex space-x-2 mt-4">
                    <button onclick="modifyBooking(${booking.id})" 
                            class="flex-1 px-3 py-2 text-sm bg-gray-100 hover:bg-gray-200 rounded-md">
                        Modify
                    </button>
                    <button onclick="cancelBooking(${booking.id})" 
                            class="flex-1 px-3 py-2 text-sm bg-red-100 hover:bg-red-200 text-red-700 rounded-md">
                        Cancel
                    </button>
                </div>
            </div>
        `;
    }).join('');
}

// ensure total_cost is included
async function confirmReservation() {
    const startTime = document.getElementById('startTime').value;
    const endTime = document.getElementById('endTime').value;
    const userId = localStorage.getItem('userId');

    if (!userId || !selectedVehicle) {
        showMessage('Invalid reservation details', false);
        return;
    }

    const totalCost = calculateRentalCost(startTime, endTime, selectedVehicle.hourly_rate || 9);

    try {
        const response = await fetch(`${SUPABASE_URL}/rest/v1/bookings`, {
            method: 'POST',
            headers: {
                'apikey': SUPABASE_KEY,
                'Authorization': `Bearer ${SUPABASE_KEY}`,
                'Content-Type': 'application/json',
                'Prefer': 'return=minimal'
            },
            body: JSON.stringify({
                user_id: userId,
                vehicle_id: selectedVehicle.id,
                start_time: startTime,
                end_time: endTime,
                status: 'pending',
                total_cost: parseFloat(totalCost) // Ensure it's stored as a number
            })
        });

        if (!response.ok) throw new Error('Failed to create booking');

        closeModal('confirmationModal');
        showMessage('Booking created successfully!', true);
        await loadMyBookings();
        await loadAvailableVehicles();
    } catch (error) {
        console.error('Error:', error);
        showMessage('Failed to create booking', false);
    }
}

// handle total_cost
async function modifyBooking(bookingId) {
    const startTime = prompt('Enter new start time (YYYY-MM-DD HH:MM):', '2024-12-07 10:00');
    const endTime = prompt('Enter new end time (YYYY-MM-DD HH:MM):', '2024-12-07 18:00');

    if (!startTime || !endTime) return;

    try {
        const bookingResponse = await fetch(
            `${SUPABASE_URL}/rest/v1/bookings?id=eq.${bookingId}&select=*,vehicle:vehicles(*)`,
            {
                headers: {
                    'apikey': SUPABASE_KEY,
                    'Authorization': `Bearer ${SUPABASE_KEY}`
                }
            }
        );

        if (!bookingResponse.ok) throw new Error('Failed to fetch booking details');
        const [booking] = await bookingResponse.json();
        
        const newTotalCost = calculateRentalCost(startTime, endTime, booking.vehicle.hourly_rate || 9);

        const response = await fetch(`${SUPABASE_URL}/rest/v1/bookings?id=eq.${bookingId}`, {
            method: 'PATCH',
            headers: {
                'apikey': SUPABASE_KEY,
                'Authorization': `Bearer ${SUPABASE_KEY}`,
                'Content-Type': 'application/json',
                'Prefer': 'return=minimal'
            },
            body: JSON.stringify({
                start_time: startTime,
                end_time: endTime,
                total_cost: parseFloat(newTotalCost), 
                updated_at: new Date().toISOString() 
            })
        });

        if (!response.ok) throw new Error('Failed to update booking');

        showMessage('Booking updated successfully!', true);
        await loadMyBookings();
    } catch (error) {
        console.error('Error:', error);
        showMessage('Failed to update booking', false);
    }
}


async function cancelBooking(bookingId) {
    // Show confirmation dialog with more details
    const confirmCancel = confirm('Are you sure you want to cancel this booking? This action cannot be undone.');
    if (!confirmCancel) return;

    // Show loading state
    const loadingMessage = document.getElementById('responseMessage');
    loadingMessage.className = 'info-message';
    loadingMessage.textContent = 'Cancelling booking...';
    loadingMessage.classList.remove('hidden');

    try {
        // Verify auth token
        const authToken = localStorage.getItem('authToken');
        if (!authToken) {
            throw new Error('You must be logged in to cancel bookings');
        }

        // First, verify the booking exists and belongs to the user
        const userId = localStorage.getItem('userId');
        const bookingCheckResponse = await fetch(
            `${SUPABASE_URL}/rest/v1/bookings?id=eq.${bookingId}&user_id=eq.${userId}`,
            {
                headers: {
                    'apikey': SUPABASE_KEY,
                    'Authorization': `Bearer ${SUPABASE_KEY}`
                }
            }
        );

        const bookingData = await bookingCheckResponse.json();
        if (!bookingData || bookingData.length === 0) {
            throw new Error('Booking not found or unauthorized');
        }

        // Proceed with cancellation
        const response = await fetch(`${SUPABASE_URL}/rest/v1/bookings?id=eq.${bookingId}`, {
            method: 'DELETE',
            headers: {
                'apikey': SUPABASE_KEY,
                'Authorization': `Bearer ${SUPABASE_KEY}`,
                'Prefer': 'return=minimal'
            }
        });

        if (!response.ok) {
            throw new Error(`Failed to cancel booking: ${response.statusText}`);
        }

        // Show success message
        showMessage('Booking cancelled successfully!', true);
        
        // Refresh the bookings list and available vehicles
        await Promise.all([
            loadMyBookings(),
            loadAvailableVehicles()
        ]);

    } catch (error) {
        console.error('Error cancelling booking:', error);
        showMessage(error.message || 'Failed to cancel booking', false);
    }
}

async function updateBookingStatus(bookingId, status) {
    const response = await fetch(`${SUPABASE_URL}/rest/v1/bookings?id=eq.${bookingId}`, {
        method: 'PATCH',
        headers: {
            'apikey': SUPABASE_KEY,
            'Authorization': `Bearer ${SUPABASE_KEY}`,
            'Content-Type': 'application/json',
            'Prefer': 'return=minimal'
        },
        body: JSON.stringify({
            status: status,
            updated_at: new Date().toISOString()
        })
    });

    if (!response.ok) {
        throw new Error('Failed to update booking status');
    }
}

function showVehicleDetails(vehicleId) {
    selectedVehicle = vehicles.find(v => v.id === vehicleId);
    if (!selectedVehicle) {
        showMessage('Vehicle not found', false);
        return;
    }

    const modal = document.getElementById('vehicleModal');
    const details = document.getElementById('vehicleDetails');

    details.innerHTML = `
        <h4 class="text-xl font-semibold">${selectedVehicle.model}</h4>
        <div class="grid grid-cols-2 gap-4 mt-4">
            <div>
                <p class="text-sm text-gray-600">License Plate</p>
                <p class="font-medium">${selectedVehicle.license_plate}</p>
            </div>
            <div>
                <p class="text-sm text-gray-600">Type</p>
                <p class="font-medium">${selectedVehicle.type}</p>
            </div>
            <div>
                <p class="text-sm text-gray-600">Battery Level</p>
                <p class="font-medium">${selectedVehicle.battery_level}%</p>
            </div>
            <div>
                <p class="text-sm text-gray-600">Location</p>
                <p class="font-medium">${selectedVehicle.location}</p>
            </div>
            <div>
                <p class="text-sm text-gray-600">Cleanliness</p>
                <p class="font-medium capitalize">${selectedVehicle.cleanliness_status.replace('_', ' ')}</p>
            </div>
            <div>
                <p class="text-sm text-gray-600">Status</p>
                <p class="font-medium capitalize">${selectedVehicle.status}</p>
            </div>
            <div>
                <p class="text-sm text-gray-600">Hourly Rate</p>
                <p class="font-medium">$${selectedVehicle.hourly_rate || '9'}/hour</p>
            </div>
        </div>
    `;

    modal.style.display = 'flex';
}

function showReservationConfirmation(vehicleId) {
    const vehicle = vehicles.find(v => v.id === vehicleId);
    if (!vehicle) {
        showMessage('Vehicle not found', false);
        return;
    }

    selectedVehicle = vehicle;
    const modal = document.getElementById('confirmationModal');
    const details = document.getElementById('confirmationDetails');
    const startTime = document.getElementById('startTime').value;
    const endTime = document.getElementById('endTime').value;
    const rentalCost = calculateRentalCost(startTime, endTime, vehicle.hourly_rate || 9);

    details.innerHTML = `
        <div class="space-y-4">
            <div>
                <h4 class="text-lg font-semibold">${vehicle.model}</h4>
                <p class="text-sm text-gray-600">${vehicle.type} - ${vehicle.license_plate}</p>
            </div>
            <div class="space-y-2">
                <div>
                    <p class="text-sm text-gray-600">Start Time</p>
                    <p class="font-medium">${new Date(startTime).toLocaleString()}</p>
                </div>
                <div>
                    <p class="text-sm text-gray-600">End Time</p>
                    <p class="font-medium">${new Date(endTime).toLocaleString()}</p>
                </div>
                <div class="mt-4 bg-gray-50 p-4 rounded-lg">
                    <p class="text-sm text-gray-600">Rental Cost Breakdown</p>
                    <p class="font-medium">$${vehicle.hourly_rate || 9}/hour × ${
                        Math.ceil((new Date(endTime) - new Date(startTime)) / (1000 * 60 * 60))
                    } hours</p>
                    <p class="text-lg font-bold text-blue-600">Total: $${rentalCost}</p>
                </div>
            </div>
            <div>
                <p class="text-sm text-gray-600">Location</p>
                <p class="font-medium">${vehicle.location}</p>
            </div>
            <div class="bg-blue-50 p-4 rounded">
                <p class="text-sm text-blue-800">
                    Please confirm your reservation details. Once confirmed, you'll receive a notification with access instructions.
                </p>
            </div>
        </div>
    `;

    modal.style.display = 'flex';
}

function closeModal(modalId) {
    document.getElementById(modalId).style.display = 'none';
    if (modalId === 'confirmationModal') {
        selectedVehicle = null;
    }
}

// Initialize on page load
window.addEventListener('load', () => {
    if (!localStorage.getItem('authToken')) {
        const loginButton = document.querySelector('[data-form="login"]');
        if (loginButton) {
            loginButton.classList.add('active');
            showForm('login');
        }
    }
});