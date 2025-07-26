import {useContext, useEffect, useState} from "react";
import {Box, Button, ButtonGroup, TextField} from "@mui/material";
import {MacrosAction} from "./MacrosAction.jsx";
import {Condition} from "../../Contition.jsx";
import {savePreset} from "../../api.js";
import {NotificationContext} from "../Alert/NotificationContext.jsx";

const INPUT_COUNT = 20;
const onChangeBinding = (event) => {
    let combo = '';
    if (event.ctrlKey) combo += 'ctrl+';
    if (event.shiftKey) combo += 'shift+';
    if (event.altKey) combo += 'alt+';
    if (event.metaKey) combo += 'meta+';
    combo += event.key.toLowerCase();
    if (combo === 'escape') {
        combo = 'esc';
    }
    event.target.value = combo;
    event.preventDefault();
    return false;
}

const renderItems = ({id, items = []}, conditions, setConditions) => {
    const result = []
    for (let i = 0; i < INPUT_COUNT; i++) {
        let preparedConditions = {rules: []};
        if (items.length && items[i] && (items[i]?.conditions_combinator || '') !== '') {
            preparedConditions.rules = items[i]['Conditions'].flatMap((item, index) => index < items[i]['Conditions'].length - 1 ? [item, items[i]?.conditions_combinator] : [item]);
            preparedConditions.rules = items[i]['Conditions'];
        } else {
            preparedConditions.rules = !items.length ? [] : items[i]?.Conditions || []
        }
        console.log('rerender');
        // const preparedConditions = {rules: []}
        result.push(<Box sx={{display: 'flex', gap: 2, m: 2}} key={i}>
            <MacrosAction name={'actions[]'} initValue={!items.length ? '' : items[i]?.Action || ''}/>
            <TextField variant={"outlined"} name={'bindings[]'} label="Binding"
                       slotProps={{inputLabel: {shrink: true}}}
                       key={'binding_' + id}
                       onKeyDown={onChangeBinding} defaultValue={!items.length ? '' : items[i]?.Binding || ''}
            />
            <TextField variant={"outlined"} name={'period_milliseconds[]'} label={"Period"}
                       key={'period_' + id}
                       slotProps={{inputLabel: {shrink: true}}}
                       defaultValue={!items.length ? '' : items[i]?.period_milliseconds}
            />
            <TextField variant={"outlined"} name={'delay_milliseconds[]'} label={"Delay"}
                       key={'delay_' + id}
                       slotProps={{inputLabel: {shrink: true}}}
                       defaultValue={!items.length ? '' : items[i]?.delay_milliseconds}
            />
            <TextField variant={"outlined"} name={'additional[]'} label={"Additional"}
                       slotProps={{inputLabel: {shrink: true}}}
                       key={'additional_' + id}
                       defaultValue={!items.length ? '' : items[i]?.Additional}
            />
            <Condition conditions={preparedConditions} fullWidth={true}
                       onQueryChange={(data) => {
                           conditions[i] = data;
                           setConditions(conditions);
                       }} idx={i}/>
        </Box>);
    }

    return result;
}
export const Macros = ({presetId, loadPresets, presetName, data = []}) => {
    const {setAlert, setSuccess} = useContext(NotificationContext);
    const [submitDisabled, disableSubmit] = useState(false);
    const [formItems, setFormItems] = useState([]);
    const [conditions, setConditions] = useState([]);
    useEffect(() => {
        console.log(data);
        setFormItems(renderItems(data, conditions, setConditions));
    }, [presetId]);
    // useEffect(() => {
    //     setFormItems(renderItems(formItemsData, setConditions));
    // }, [formItemsData, setConditions]);
    // useEffect(() => {
    //     async function initProfile() {
    //         try {
    //             const data = await getProfile(profileName);
    //             if (data) {
    //                 setFormItemsData(data);
    //                 setFormItems(renderItems(data, conditions, setConditions));
    //             } else {
    //                 setFormItems(renderItems([], conditions, setConditions));
    //             }
    //         } catch (error) {
    //             setFormItems(renderItems([], conditions, setConditions));
    //         }
    //     }
    //
    //     initProfile();
    // }, [profileName]);
    const handleSubmit = async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const obj = {items: [], name: presetName};
        for (let i = 0; i < INPUT_COUNT; i++) {
            obj.items.push({
                'action': formData.getAll('actions[]')[i],
                'binding': formData.getAll('bindings[]')[i],
                'delay_milliseconds': parseInt(formData.getAll('delay_milliseconds[]')[i]),
                'period_milliseconds': parseInt(formData.getAll('period_milliseconds[]')[i]),
                'additional': formData.getAll('additional[]')[i],
                'conditions': conditions[i].filter(item => typeof item === 'object'),
                'conditions_combinator': conditions[i].filter(item => typeof item === 'string')[0] || "",
            })
        }

        disableSubmit(true);
        try {
            await savePreset(presetId, obj);
        } catch (error) {
            setAlert(error.message);
        }
        disableSubmit(false);
        setSuccess('Saved');
        loadPresets();
    }
    return (<Box>
        <form onSubmit={handleSubmit}>
            <Box sx={{m: 2, display: 'flex'}} alignItems={'center'}>
                <Button type={"submit"} disabled={submitDisabled}>Save</Button>
            </Box>
            {formItems}
        </form>
    </Box>)
        ;
}
