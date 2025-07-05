import axios from 'axios';

const api = axios.create({
    baseURL: `http://${import.meta.env.VITE_SERVER_DOMAIN}:${import.meta.env.VITE_SERVER_PORT}/api`,
});

export const getProfile = (profileName) => {
    return api.get('/profile/' + profileName).then((response) => {
        return response.data;
    }).catch((error) => {
        console.log(error);
    })
}

export const saveProfile = (profileName, formData) => {
    return api.post('/profile/' + profileName, formData).then((response) => {
        console.log(response)
    })
}

export const startMacros = (profileName) => {
    return api.post('/start/' + profileName).then((response) => {
        console.log(response)
    })
}

export const stopMacros = () => {
    return api.post('/stop').then((response) => {
        console.log(response)
    })
}
