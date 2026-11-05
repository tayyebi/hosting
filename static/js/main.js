function togglePassword() {
    const input = document.getElementById('password');
    const eyeShow = document.getElementById('eye-show');
    const eyeHide = document.getElementById('eye-hide');

    if (input.type === 'password') {
        input.type = 'text';
        eyeShow.style.display = 'none';
        eyeHide.style.display = 'block';
    } else {
        input.type = 'password';
        eyeShow.style.display = 'block';
        eyeHide.style.display = 'none';
    }
}

function toggleWeb3() {
    const toggle = document.querySelector('.web3-toggle');
    const wallets = document.getElementById('web3-wallets');
    const expanded = toggle.getAttribute('aria-expanded') === 'true';

    toggle.setAttribute('aria-expanded', String(!expanded));
    if (expanded) {
        wallets.hidden = true;
    } else {
        wallets.hidden = false;
    }
}

// Subtle submit animation
document.addEventListener('DOMContentLoaded', () => {
    const form = document.querySelector('.auth-form');
    const btn = form && form.querySelector('.btn--primary');

    if (form && btn) {
        form.addEventListener('submit', () => {
            btn.disabled = true;
            btn.querySelector('.btn-text').textContent = 'Signing in…';
        });
    }
});
