function uploadFile() {
    const fileInput = document.getElementById('fileInput');
    if (!fileInput.files.length) {
        alert('Please select a file.');
        return;
    }

    const formData = new FormData();
    formData.append('file', fileInput.files[0]);

    fetch('/upload', {
        method: 'POST',
        body: formData,
    })
    .then(response => {
        if (response.ok) {
            document.getElementById('uploadStatus').innerText = 'Upload successful!';
            listFiles(); // Refresh the file list
        } else {
            throw new Error('Failed to upload file.');
        }
    })
    .catch(error => {
        document.getElementById('uploadStatus').innerText = error.message;
        console.error('Error:', error);
    });
}

function listFiles() {
    fetch('/files')
    .then(response => response.json())
    .then(files => {
        const fileList = document.getElementById('fileList');
        fileList.innerHTML = ''; // Clear current list
        files.forEach(file => {
            const li = document.createElement('li');
            li.textContent = file;
            li.onclick = () => downloadFile(file);
            fileList.appendChild(li);
        });
    })
    .catch(error => {
        console.error('Error:', error);
    });
}

function downloadFile(filename) {
    window.location.href = `/files/${filename}`;
}
