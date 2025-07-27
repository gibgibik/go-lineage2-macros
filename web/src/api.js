import axios from 'axios';

const api = axios.create({
    baseURL: `http://${import.meta.env.VITE_SERVER_DOMAIN}:${import.meta.env.VITE_SERVER_PORT}/api`,
});

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

export const pauseMacros = (pid) => {
    return api.post('/pause', {pid}).then((response) => {
    })
}

export const getPresetsList = () => {
    return api.get('/preset');
}

export const getProfilesList = () => {
    return api.get('/profile?v=1');
}

export const savePreset = (presetId, data) => {
    return api.post('/preset/' + presetId, data).then((response) => {
    })
}
