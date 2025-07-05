import {Box, Button, ButtonGroup, Container, TextField} from "@mui/material";
import {MacrosAction} from "./MacrosAction.jsx";
import {getProfile, saveProfile} from "./api.js";
import {useEffect, useState} from "react";
import {isHotkeyPressed, useHotkeys} from "react-hotkeys-hook";

const INPUT_COUNT = 10;
const PROFILE_NAME = 'static_profile';//@make few profiles?
const onChangeBinding = (event) => {
    let combo = '';
    if (event.ctrlKey) combo += 'ctrl+';
    if (event.shiftKey) combo += 'shift+';
    if (event.altKey) combo += 'alt+';
    if (event.metaKey) combo += 'meta+';
    combo += event.key.toLowerCase();
    if (combo === 'escape') {
        return true;
    }
    event.target.value = combo;
    event.preventDefault();
    return false;
}
const renderItems = ({'Actions': actions = [], 'Bindings': bindings = [], 'Period_seconds': periodSeconds = []}) => {
    const items = []
    for (let i = 0; i < INPUT_COUNT; i++) {
        items.push(<Box sx={{display: 'flex', gap: 2, m: 2}} key={i}>
            <MacrosAction name={'actions[]'} initValue={actions[i] || ''}/>
            <TextField variant={"outlined"} fullWidth={true} name={'bindings[]'} onKeyDown={onChangeBinding} defaultValue={bindings[i] || ''}
            />
            <TextField variant={"outlined"} fullWidth={true} name={'period_seconds[]'}
                       defaultValue={periodSeconds[i] || ''}
            />
        </Box>);
    }
    console.log('render');

    return items;
}
export const Macros = () => {
    const [submitDisabled, disableSubmit] = useState(false);
    const [formItemsData, setFormItemsData] = useState({Actions: [], Bindings: [], Period_seconds: []});
    const [formItems, setFormItems] = useState([]);
    useEffect(() => {
        setFormItems(renderItems(formItemsData));
    }, [formItemsData]);
    useEffect(() => {
        async function initProfile() {
            try {
                const data = await getProfile(PROFILE_NAME);
                if (data) {
                    setFormItemsData(data);
                }
            } catch (error) {
            }
        }

        initProfile();
    }, []);
    useEffect(() => {
        if (!formItemsData || !Object.keys(formItemsData).length) {
            return;
        }
        console.log(formItemsData);
    }, [formItemsData])
    const handleSubmit = async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        // const data = Object.fromEntries(formData.entries());
        // console.log(Object.fromEntries(formData.entries()))
        const obj = {};
        for (let [key, value] of formData.entries()) {
            if (key === 'period_seconds[]') {
                value = parseInt(value);
            }
            if (key.endsWith('[]')) {
                const cleanKey = key.slice(0, -2);
                if (!obj[cleanKey]) obj[cleanKey] = [];
                obj[cleanKey].push(value);
            } else {
                obj[key] = value;
            }
        }
        obj['profile'] = PROFILE_NAME

        disableSubmit(true);
        await saveProfile(PROFILE_NAME, obj);
        disableSubmit(false);
    }
    return (<Container>
        <form onSubmit={handleSubmit}>
            {formItems}
            <ButtonGroup variant="contained" sx={{gap: 4, display: 'flex', justifyContent: 'center'}}>
                <Button type={"submit"} disabled={submitDisabled}>Save</Button>
            </ButtonGroup>
        </form>
    </Container>)
        ;
}
