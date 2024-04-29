function togglePasswordVisibility(id) {
    var passwordInput = document.getElementById(id);
    if (passwordInput.type === "password") {
        passwordInput.type = "text";
    } else {
        passwordInput.type = "password";
    }
}

function validateForm(event) {
    event.preventDefault(); // Empêche la soumission normale du formulaire

    var password = document.getElementById("password").value;
    var confirmPassword = document.getElementById("confirm_password").value;

    // Vérifie si l'utilisateur a cliqué sur le bouton "Register"
    if (event.submitter.id === 'register-button') {
        if (password != confirmPassword) {
            alert("Erreur : Les mots de passe ne correspondent pas."); // Modifiez le message ici
            return false;
        }
    }

    // Crée une nouvelle requête AJAX
    var xhr = new XMLHttpRequest();
    xhr.open("POST", "/signup", true);
    xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");

    // Gère la réponse
    xhr.onreadystatechange = function() {
        if (this.readyState === XMLHttpRequest.DONE && this.status === 200) {
            // Traite la réponse ici. Par exemple, vous pouvez rediriger l'utilisateur vers une autre page
            window.location.href = "/some-other-page";
        }
    }

    // Envoie la requête
    var formData = new FormData(event.target);
    xhr.send(new URLSearchParams(new FormData(formData)).toString());

    return false;
}

// Attachez la fonction de validation au formulaire lors du chargement de la page
window.onload = function() {
    var form = document.querySelector('form[action="/signup"]');
    if (form) {
        form.onsubmit = validateForm;
    }
}

function toggleDropdown() {
    const togglemenu = documentquerySelector('.dropdown-content');
    togglemenu.classList.toggle('active');
}