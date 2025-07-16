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

export const startMacros = (profileName, pid) => {
    return api.post('/start/' + profileName, {pid}).then((response) => {
    })
}

export const stopMacros = (pid) => {
    return api.post('/stop', {pid}).then((response) => {
    })
}

export const init = () => {
    return api.get('/init');
}
