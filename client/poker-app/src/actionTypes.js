import { set } from 'lodash'

/**
 * List of action type names.
 *
 * Action type names are namespaced by section and view.
 */
const actionNames = [
  'CHAT.NEW_MESSAGE',
  'GAME.UPDATE',
  'SERVER.ERROR',
  'SERVER.ON_JOIN',
  'SERVER.ON_TAKE_SEAT',
  'SERVER.JOIN',
  'WEBRTC.NEW_PEER',
  'WEBRTC.ON_RECEIVE_ANSWER',
  'WEBRTC.ON_RECEIVE_OFFER',
]

/**
 * Conversion of action names to a nested object, allowing dot-notation access for action types, e.g.
 *
 *   case actionTypes.SECTION.VIEW.ACTION
 */
const actionTypes = actionNames.reduce((acc, name) => set(acc, name, name), {})

export default actionTypes
