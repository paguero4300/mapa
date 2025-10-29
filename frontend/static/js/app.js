// Funciones utilitarias para la aplicación

// Utilidad para hacer peticiones fetch con manejo de errores
async function apiRequest(url, options = {}) {
    const defaultOptions = {
        credentials: 'include', // Importante: incluir cookies automáticamente
        headers: {
            'Content-Type': 'application/json',
        },
    };
    
    const finalOptions = {
        ...defaultOptions,
        ...options,
        headers: {
            ...defaultOptions.headers,
            ...options.headers,
        },
    };
    
    try {
        const response = await fetch(url, finalOptions);
        
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || `HTTP ${response.status}`);
        }
        
        return await response.json();
    } catch (error) {
        console.error('API Request Error:', error);
        throw error;
    }
}

// Utilidad para formatear fechas
function formatDate(dateString) {
    if (!dateString) return 'N/A';
    
    const date = new Date(dateString);
    return date.toLocaleString('es-ES', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
    });
}

// Utilidad para mostrar notificaciones
function showNotification(message, type = 'info') {
    // Crear elemento de notificación
    const notification = document.createElement('div');
    notification.className = `fixed top-4 right-4 p-4 rounded-lg shadow-lg z-50 fade-in ${
        type === 'error' ? 'bg-red-500 text-white' :
        type === 'success' ? 'bg-green-500 text-white' :
        type === 'warning' ? 'bg-yellow-500 text-black' :
        'bg-blue-500 text-white'
    }`;
    notification.textContent = message;
    
    document.body.appendChild(notification);
    
    // Remover después de 3 segundos
    setTimeout(() => {
        notification.remove();
    }, 3000);
}

// Utilidad para validar email
function isValidEmail(email) {
    const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return re.test(email);
}

// Verificar si el usuario está autenticado
async function isAuthenticated() {
    try {
        await apiRequest('/api/session');
        return true;
    } catch (error) {
        return false;
    }
}

// Redirigir si no está autenticado
async function requireAuth() {
    if (!(await isAuthenticated())) {
        window.location.href = '/';
        return false;
    }
    return true;
}

// Utilidad para manejo de errores global
window.addEventListener('unhandledrejection', (event) => {
    console.error('Unhandled promise rejection:', event.reason);
    showNotification('Ocurrió un error inesperado', 'error');
});

// Utilidad para detectar conexión
function isOnline() {
    return navigator.onLine;
}

// Event listener para cambios de conexión
window.addEventListener('online', () => {
    showNotification('Conexión restaurada', 'success');
});

window.addEventListener('offline', () => {
    showNotification('Sin conexión a internet', 'warning');
});

// Exportar funciones para uso global
window.appUtils = {
    apiRequest,
    formatDate,
    showNotification,
    isValidEmail,
    isOnline,
    isAuthenticated,
    requireAuth,
};

// ALPINE.JS STORE DESHABILITADO - USANDO SOLO JAVASCRIPT VANILLA
// El store de Alpine.js causaba requests innecesarios y confusión
console.log('🚫 [ALPINE] Store de Alpine.js deshabilitado - usando JavaScript vanilla');

// DOM READY - SOLO PARA INICIALIZACIÓN BÁSICA
document.addEventListener('DOMContentLoaded', () => {
    console.log('🌐 [DOM] DOM cargado - JavaScript vanilla inicializado');
    
    // Solo verificar si estamos en una página que requiere verificación
    const currentPath = window.location.pathname;
    if (currentPath === '/' || currentPath === '/login') {
        console.log('📝 [DOM] Página de login - no verificar sesión automáticamente');
        return;
    }
    
    // Solo verificar sessionStorage en páginas protegidas
    if (currentPath.startsWith('/dashboard')) {
        const savedUserData = sessionStorage.getItem('userData');
        if (savedUserData) {
            try {
                const userData = JSON.parse(savedUserData);
                console.log('💾 [DOM] Datos encontrados en sessionStorage para dashboard:', userData.name);
            } catch (error) {
                console.error('❌ [DOM] Error al recuperar datos de sessionStorage:', error);
                sessionStorage.removeItem('userData');
            }
        }
    }
});

// Inicialización cuando el DOM está listo
document.addEventListener('DOMContentLoaded', () => {
    console.log('Traccar Login App initialized');
    
    // Verificar conexión inicial
    if (!isOnline()) {
        showNotification('Sin conexión a internet', 'warning');
    }
});