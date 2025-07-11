import axios from 'axios';

const api = axios.create({
    baseURL: `http://${import.meta.env.VITE_SERVER_DOMAIN}:${import.meta.env.VITE_SERVER_PORT}/api`,
});

export const getProfile = (profileName) => {
    return api.get('/profile/' + profileName).then((response) => {
        return response.data;
    }).catch((error) => {
    })
}

export const saveProfile = (profileName, formData) => {
    return api.post('/profile/' + profileName, formData).then((response) => {
    })
}

export const startMacros = (profileName) => {
    return api.post('/start/' + profileName).then((response) => {
    })
}

export const stopMacros = () => {
    return api.post('/stop').then((response) => {
    })
}

export const init = () => {
    return api.get('/init');
}
