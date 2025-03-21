document.addEventListener('DOMContentLoaded', function() {
    // Setup WebSocket connection
    setupWebSocket();
    
    // Setup command selection logic
    document.getElementById('command-select').addEventListener('change', function() {
        const command = this.value;
        // Show/hide dig options
        document.getElementById('dig-options-container').style.display = 
            command === 'dig' ? 'block' : 'none';
        
        // Show/hide iperf options
        document.getElementById('iperf-container').style.display = 
            command === 'iperf3' ? 'block' : 'none';
    });
    
    // Setup clear output button
    document.getElementById('clear-output').addEventListener('click', function() {
        document.getElementById('output').textContent = '';
    });
});

let socket;

function setupWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    socket = new WebSocket(wsUrl);
    
    socket.onopen = function() {
        console.log('WebSocket connection established');
    };
    
    socket.onmessage = function(event) {
        const output = document.getElementById('output');
        output.textContent += event.data;
        output.scrollTop = output.scrollHeight;
    };
    
    socket.onclose = function() {
        console.log('WebSocket connection closed');
        // Try to reconnect after a delay
        setTimeout(setupWebSocket, 3000);
    };
    
    socket.onerror = function(error) {
        console.error('WebSocket error:', error);
    };
    
    // Setup run command button
    document.getElementById('run-command').addEventListener('click', function() {
        runCommand();
    });
}

function runCommand() {
    if (socket.readyState !== WebSocket.OPEN) {
        alert('WebSocket connection is not open. Please refresh the page.');
        return;
    }
    
    const command = document.getElementById('command-select').value;
    const target = document.getElementById('target').value.trim();
    
    if (!command) {
        alert('Please select a command');
        return;
    }
    
    if (!target && command !== 'ifconfig') {
        alert('Please enter a target hostname or IP address');
        return;
    }
    
    const output = document.getElementById('output');
    output.textContent = '';
    
    let option = '';
    let parallel = '';
    let duration = '';
    
    if (command === 'dig') {
        // Get the single selected option value
        let digOptions = document.getElementById('dig-options');
        option = digOptions.options[digOptions.selectedIndex].value;
    } else if (command === 'iperf3') {
        parallel = document.getElementById('iperf-parallel').value;
        duration = document.getElementById('iperf-duration').value;
    }
    
    const request = {
        command: command,
        target: target,
        option: option,
        parallel: parallel,
        duration: duration
    };
    
    socket.send(JSON.stringify(request));
}