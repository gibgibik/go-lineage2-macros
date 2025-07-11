import {Box, Button, ButtonGroup, TextField} from "@mui/material";
import {MacrosAction} from "./MacrosAction.jsx";
import {getProfile, saveProfile} from "./api.js";
import React, {useEffect, useState} from "react";
import {Condition} from "./Contition.jsx";

const INPUT_COUNT = 10;
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

const renderItems = ({Items: items = []}, conditions, setConditions) => {
    const result = []
    for (let i = 0; i < INPUT_COUNT; i++) {
        result.push(<Box sx={{display: 'flex', gap: 2, m: 2}} key={i}>
            <MacrosAction name={'actions[]'} initValue={!items.length ? '' : items[i]['Action']}/>
            <TextField variant={"outlined"} name={'bindings[]'} label="Binding"
                       slotProps={{inputLabel: {shrink: true}}}
                       onKeyDown={onChangeBinding} defaultValue={!items.length ? '' : items[i]['Binding']}
            />
            <TextField variant={"outlined"} name={'period_seconds[]'} label={"Period Seconds"}
                       slotProps={{inputLabel: {shrink: true}}}
                       defaultValue={!items.length ? '' : items[i]['period_seconds']}
            />
            <TextField variant={"outlined"} name={'additional[]'} label={"Additional"}
                       slotProps={{inputLabel: {shrink: true}}}
                       defaultValue={!items.length ? '' : items[i]['Additional']}
            />
            <Condition conditions={{rules: !items.length ? [] : items[i]['Conditions']}} fullWidth={true}
                       onQueryChange={(data) => {
                           conditions[i] = data;
                           setConditions(conditions);
                       }} idx={i}/>
        </Box>);
    }

    return result;
}
export const Macros = ({profileName}) => {
    const [submitDisabled, disableSubmit] = useState(false);
    const [formItemsData, setFormItemsData] = useState([]);
    const [formItems, setFormItems] = useState([]);
    const [conditions, setConditions] = useState([]);
    // useEffect(() => {
    //     setFormItems(renderItems(formItemsData, setConditions));
    // }, [formItemsData, setConditions]);
    useEffect(() => {
        async function initProfile() {
            try {
                const data = await getProfile(profileName);
                if (data) {
                    setFormItemsData(data);
                    setFormItems(renderItems(data, conditions, setConditions));
                } else {
                    setFormItems(renderItems([], conditions, setConditions));
                }
            } catch (error) {
                setFormItems(renderItems([], conditions, setConditions));
            }
        }

        initProfile();
    }, []);
    useEffect(() => {
        if (!formItemsData || !Object.keys(formItemsData).length) {
            return;
        }
    }, [formItemsData])
    const handleSubmit = async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const obj = {items: [], profile: profileName};
        for (let i = 0; i < INPUT_COUNT; i++) {
            obj.items.push({
                'action': formData.getAll('actions[]')[i],
                'binding': formData.getAll('bindings[]')[i],
                'period_seconds': parseInt(formData.getAll('period_seconds[]')[i]),
                'additional': formData.getAll('additional[]')[i],
                'conditions': conditions[i],
            })
        }

        // console.log(conditions);
        // console.log(obj);
        // return;
        disableSubmit(true);
        await saveProfile(profileName, obj);
        disableSubmit(false);
    }
    return (<Box>
        <form onSubmit={handleSubmit}>
            {formItems}
            <ButtonGroup variant="contained" sx={{gap: 4, display: 'flex', justifyContent: 'center'}}>
                <Button type={"submit"} disabled={submitDisabled}>Save</Button>
            </ButtonGroup>
        </form>
    </Box>)
        ;
}
