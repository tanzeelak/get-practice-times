// Use current origin for production, localhost for development
const API_BASE_URL = window.location.hostname === 'localhost' ? 'http://localhost:8080' : window.location.origin;

let scheduleData = null;

// Studio type classifications
const PRIVATE_STUDIOS = ['Studio 1', 'Studio 2', 'Studio 3', 'Studio 4', 'Studio 8', 'Studio 10'];
const GROUP_STUDIOS = ['Studio 5', 'Studio 9', 'Studio B', 'Studio C', 'Studio D', 'Studio E'];

function getStudioType(studioName) {
    if (PRIVATE_STUDIOS.includes(studioName)) return 'private';
    if (GROUP_STUDIOS.includes(studioName)) return 'group';
    return 'private'; // default fallback
}

// Initialize the app
document.addEventListener('DOMContentLoaded', function() {
    fetchRehearsals();
});

async function fetchRehearsals() {
    const refreshBtn = document.getElementById('refreshBtn');
    const refreshText = document.getElementById('refreshText');
    const loadingText = document.getElementById('loadingText');
    const errorMessage = document.getElementById('error-message');
    const scheduleContainer = document.getElementById('schedule-container');

    // Show loading state
    refreshBtn.disabled = true;
    refreshText.classList.add('hidden');
    loadingText.classList.remove('hidden');
    errorMessage.classList.add('hidden');
    
    // Show loading spinner
    scheduleContainer.innerHTML = `
        <div class="loading">
            <div class="spinner"></div>
            <p>Loading rehearsal availability...</p>
        </div>
    `;

    try {
        const response = await fetch(`${API_BASE_URL}/api/rehearsals`);
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        scheduleData = data;
        
        renderSchedule(data);
        
    } catch (error) {
        console.error('Error fetching rehearsal data:', error);
        showError(`Failed to load rehearsal data: ${error.message}`);
    } finally {
        // Reset button state
        refreshBtn.disabled = false;
        refreshText.classList.remove('hidden');
        loadingText.classList.add('hidden');
    }
}

function showError(message) {
    const errorMessage = document.getElementById('error-message');
    const scheduleContainer = document.getElementById('schedule-container');
    
    errorMessage.textContent = message;
    errorMessage.classList.remove('hidden');
    
    scheduleContainer.innerHTML = `
        <div style="text-align: center; padding: 40px; color: #666;">
            <p>Unable to load rehearsal data. Please try refreshing.</p>
        </div>
    `;
}

function renderSchedule(data) {
    const scheduleContainer = document.getElementById('schedule-container');
    
    if (!data || Object.keys(data).length === 0) {
        scheduleContainer.innerHTML = `
            <div style="text-align: center; padding: 40px; color: #666;">
                <p>No rehearsal slots available at this time.</p>
            </div>
        `;
        return;
    }

    let html = '';
    
    Object.entries(data).forEach(([date, timeSlots], index) => {
        const isFirstDay = index === 0;
        const dayId = `day-${index}`;
        
        html += `
            <div class="day-card">
                <div class="day-header ${isFirstDay ? 'expanded' : ''}" onclick="toggleDay('${dayId}')">
                    <h3>${date}</h3>
                    <span class="toggle-icon">▼</span>
                </div>
                <div class="time-slots ${isFirstDay ? 'expanded' : ''}" id="${dayId}">
        `;
        
        if (Object.keys(timeSlots).length === 0) {
            html += `
                <div style="padding: 20px; text-align: center; color: #666;">
                    No available time slots for this day.
                </div>
            `;
        } else {
            Object.entries(timeSlots).forEach(([time, studios]) => {
                html += `
                    <div class="time-slot">
                        <div class="time">${time}</div>
                        <div class="studios">
                `;
                
                // Group studios by type
                const privateStudios = studios.filter(studio => getStudioType(studio) === 'private');
                const groupStudios = studios.filter(studio => getStudioType(studio) === 'group');
                
                // Display private studios first, then group studios
                privateStudios.forEach(studio => {
                    html += `<span class="studio private">${studio}</span>`;
                });
                
                groupStudios.forEach(studio => {
                    html += `<span class="studio group">${studio}</span>`;
                });
                
                html += `
                        </div>
                    </div>
                `;
            });
        }
        
        html += `
                </div>
            </div>
        `;
    });
    
    scheduleContainer.innerHTML = html;
}

function toggleDay(dayId) {
    const timeSlots = document.getElementById(dayId);
    const header = timeSlots.previousElementSibling;
    
    if (timeSlots.classList.contains('expanded')) {
        timeSlots.classList.remove('expanded');
        header.classList.remove('expanded');
    } else {
        timeSlots.classList.add('expanded');
        header.classList.add('expanded');
    }
}

// Utility function to format date nicely
function formatDate(dateString) {
    try {
        const date = new Date(dateString + ', 2024'); // Add year for parsing
        return date.toLocaleDateString('en-US', { 
            weekday: 'long', 
            month: 'long', 
            day: 'numeric' 
        });
    } catch (error) {
        return dateString; // Fallback to original string
    }
}

// Auto-refresh once a week
setInterval(() => {
    console.log('Auto-refreshing rehearsal data...');
    fetchRehearsals();
}, 7 * 24 * 60 * 60 * 1000); // 1 week

// Add keyboard shortcuts
document.addEventListener('keydown', function(event) {
    // Press 'R' to refresh
    if (event.key === 'r' || event.key === 'R') {
        if (!event.ctrlKey && !event.metaKey) { // Avoid conflicts with browser refresh
            event.preventDefault();
            fetchRehearsals();
        }
    }
});