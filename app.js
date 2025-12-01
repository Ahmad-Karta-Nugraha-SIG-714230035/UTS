// Initialize the map
const map = L.map('map').setView([-6.2, 106.816666], 10); // Default to Jakarta coordinates

// Add OpenStreetMap tiles
L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '© OpenStreetMap contributors'
}).addTo(map);

// Add click event to map to fill coordinates
map.on('click', function(e) {
    const lat = e.latlng.lat.toFixed(6);
    const lng = e.latlng.lng.toFixed(6);

    document.getElementById('lat').value = lat;
    document.getElementById('lng').value = lng;

    // Optional: Show a temporary marker at clicked location
    const tempMarker = L.marker([lat, lng]).addTo(map);
    tempMarker.bindPopup(`Clicked location:<br>Lat: ${lat}<br>Lng: ${lng}`).openPopup();

    // Remove temp marker after 3 seconds
    setTimeout(() => {
        map.removeLayer(tempMarker);
    }, 3000);
});

// Function to load features from the backend
async function loadFeatures() {
    try {
        const response = await fetch('/api/features');
        const features = await response.json();
        features.forEach(feature => {
            addMarker(feature);
        });
    } catch (error) {
        console.error('Error loading features:', error);
    }
}

// Function to add a marker to the map
function addMarker(feature) {
    const marker = L.marker([feature.lat, feature.lng]).addTo(map);
    marker.bindPopup(`<b>${feature.name}</b><br>Category: ${feature.category}`);
}

// Handle form submission
document.getElementById('featureForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const name = document.getElementById('name').value;
    const lat = parseFloat(document.getElementById('lat').value);
    const lng = parseFloat(document.getElementById('lng').value);
    const category = document.getElementById('category').value;
    
    const feature = { name, lat, lng, category };
    
    try {
        const response = await fetch('/api/features', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(feature)
        });
        
        if (response.ok) {
            // Reload all features to ensure consistency
            loadFeatures();
            document.getElementById('message').textContent = 'Feature added successfully!';
            document.getElementById('message').style.color = 'green';
            document.getElementById('featureForm').reset();
        } else {
            document.getElementById('message').textContent = 'Error adding feature.';
            document.getElementById('message').style.color = 'red';
        }
    } catch (error) {
        console.error('Error adding feature:', error);
        document.getElementById('message').textContent = 'Error adding feature.';
        document.getElementById('message').style.color = 'red';
    }
});

// Load features when the page loads
loadFeatures();
