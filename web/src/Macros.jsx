import {Box, Button, ButtonGroup, Container, TextField} from "@mui/material";
import {MacrosAction} from "./MacrosAction.jsx";
import {getProfile, saveProfile} from "./api.js";
import React, {useEffect, useState} from "react";
import {IMaskInput} from "react-imask";

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

const ConditionMask = React.forwardRef(
    function TextMaskCustom(props, ref) {
        const {onChange, ...other} = props;
        return (
            <IMaskInput
                {...other}
                mask="P\P S ##%"
                definitions={{
                    '#': /[0-9]/,
                    'P': /H|M/,
                    'S': />|<|=/,
                }}
                inputRef={ref}
                onAccept={(value) => onChange({target: {name: props.name, value}})}
                overwrite
            />
        );
    },
);

const renderItems = ({
                         'Actions': actions = [],
                         'Bindings': bindings = [],
                         'Period_seconds': periodSeconds = [],
                         'Start_target_condition': startTargetCondition = [],
                         'End_target_condition': endTargetCondition = [],
    'Use_condition': useCondition = [],
                     }) => {
    console.log(startTargetCondition, endTargetCondition);
    const items = []
    for (let i = 0; i < INPUT_COUNT; i++) {
        items.push(<Box sx={{display: 'flex', gap: 2, m: 2}} key={i}>
            <MacrosAction name={'actions[]'} initValue={actions[i] || ''}/>
            <TextField variant={"outlined"} fullWidth={true} name={'bindings[]'} label="Binding"
                       slotProps={{inputLabel: {shrink: true}}}
                       onKeyDown={onChangeBinding} defaultValue={bindings[i] || ''}
            />
            <TextField variant={"outlined"} fullWidth={true} name={'period_seconds[]'} label={"Interval Seconds"}
                       slotProps={{inputLabel: {shrink: true}}}
                       defaultValue={periodSeconds[i] || ''}
            />
            <TextField variant={"outlined"} fullWidth={true} name={'start_target_condition[]'} label={"Start Target Condition"}
                       slotProps={{inputLabel: {shrink: true}}}
                       defaultValue={startTargetCondition[i] || ''}
                       InputProps={{inputComponent:ConditionMask }}
            />
            <TextField variant={"outlined"} fullWidth={true} name={'end_target_condition[]'} label={"End Target Condition"}
                       slotProps={{inputLabel: {shrink: true}}}
                       defaultValue={endTargetCondition[i] || ''}
                       InputProps={{inputComponent:ConditionMask }}

            />
            <TextField variant={"outlined"} fullWidth={true} name={'use_condition[]'} label={"Use Condition"}
                       slotProps={{inputLabel: {shrink: true}}}
                       defaultValue={useCondition[i] || ''}
                       InputProps={{inputComponent:ConditionMask }}

            />
        </Box>);
    }

    return items;
}
export const Macros = ({profileName}) => {
    const [submitDisabled, disableSubmit] = useState(false);
    const [formItemsData, setFormItemsData] = useState({Actions: [], Bindings: [], Period_seconds: []});
    const [formItems, setFormItems] = useState([]);
    useEffect(() => {
        setFormItems(renderItems(formItemsData));
    }, [formItemsData]);
    useEffect(() => {
        async function initProfile() {
            try {
                const data = await getProfile(profileName);
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
        obj['profile'] = profileName

        disableSubmit(true);
        await saveProfile(profileName, obj);
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
