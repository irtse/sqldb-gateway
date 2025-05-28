// custom.js
window.addEventListener('load', function () {
    const toggle = document.querySelector('.sidemenu-toggle');
    if (toggle && document.body.classList.contains('sidemenu-expanded')) {
        toggle.click();
    }
});