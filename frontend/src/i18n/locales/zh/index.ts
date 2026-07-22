import landing from './landing'
import common from './common'
import dashboard from './dashboard'
import admin from './admin'
import misc from './misc'
import modelPlaza from './modelPlaza'

export default {
  ...landing,
  ...common,
  ...dashboard,
  admin,
  ...misc,
  ...modelPlaza,
}
